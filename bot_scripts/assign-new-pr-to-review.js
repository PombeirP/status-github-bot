// Description:
//   Script that listens to new GitHub pull requests
//   and assigns them to the REVIEW column on the 'Pipeline for QA' project
//
// Dependencies:
//   github: "^13.1.0"
//   probot-config: "^0.1.0"
//   probot-slack-status: "^0.2.2"
//
// Author:
//   PombeirP

const defaultConfig = require('../lib/config')

const getConfig = require('probot-config')
const Slack = require('probot-slack-status')
const slackHelper = require('../lib/slack')

const botName = 'assign-new-pr-to-review'
let slackClient = null

module.exports = (robot) => {
  // robot.on('slack.connected', ({ slack }) => {
  Slack(robot, (slack) => {
    robot.log.trace(`${botName} - Connected, assigned slackClient`)
    slackClient = slack
  })

  robot.on('pull_request.opened', async context => {
    // Make sure we don't listen to our own messages
    if (context.isBot) { return }

    // A new PR was opened
    await assignPullRequestToReview(context, robot)
  })
}

async function assignPullRequestToReview (context, robot) {
  const { github, payload } = context
  const config = await getConfig(context, 'github-bot.yml', defaultConfig(robot, '.github/github-bot.yml'))
  const ownerName = payload.repository.owner.login
  const repoName = payload.repository.name
  const prNumber = payload.pull_request.number

  const projectBoardConfig = config ? config['project-board'] : null
  if (!projectBoardConfig) {
    return
  }

  robot.log(`${botName} - Handling Pull Request #${prNumber} on repo ${ownerName}/${repoName}`)

  // Fetch repo projects
  // TODO: The repo project and project column info should be cached
  // in order to improve performance and reduce roundtrips
  let column = null
  const projectBoardName = projectBoardConfig.name
  const reviewColumnName = projectBoardConfig['review-column-name']
  try {
    const ghprojectsPayload = await github.projects.getRepoProjects({
      owner: ownerName,
      repo: repoName,
      state: 'open'
    })

    // Find 'Pipeline for QA' project
    const project = ghprojectsPayload.data.find(p => p.name === projectBoardName)
    if (!project) {
      robot.log.error(`${botName} - Couldn't find project ${projectBoardName} in repo ${ownerName}/${repoName}`)
      return
    }

    robot.log.debug(`${botName} - Fetched ${project.name} project (${project.id})`)

    // Fetch REVIEW column ID
    try {
      const ghcolumnsPayload = await github.projects.getProjectColumns({ project_id: project.id })

      column = ghcolumnsPayload.data.find(c => c.name === reviewColumnName)
      if (!column) {
        robot.log.error(`${botName} - Couldn't find ${reviewColumnName} column in project ${project.name}`)
        return
      }

      robot.log.debug(`${botName} - Fetched ${column.name} column (${column.id})`)
    } catch (err) {
      robot.log.error(`${botName} - Couldn't fetch the github columns for project: ${err}`, ownerName, repoName, project.id)
      return
    }
  } catch (err) {
    robot.log.error(`${botName} - Couldn't fetch the github projects for repo: ${err}`, ownerName, repoName)
    return
  }

  // Create project card for the PR in the REVIEW column
  try {
    if (process.env.DRY_RUN) {
      robot.log.debug(`${botName} - Would have created card`, column.id, payload.pull_request.id)
    } else {
      const ghcardPayload = await github.projects.createProjectCard({
        column_id: column.id,
        content_type: 'PullRequest',
        content_id: payload.pull_request.id
      })

      robot.log.debug(`${botName} - Created card: ${ghcardPayload.data.url}`, ghcardPayload.data.id)
    }

    // Send message to Slack
    slackHelper.sendMessage(robot, slackClient, config.slack.notification.room, `Assigned PR to ${reviewColumnName} in ${projectBoardName} project\n${payload.pull_request.html_url}`)
  } catch (err) {
    robot.log.error(`${botName} - Couldn't create project card for the PR: ${err}`, column.id, payload.pull_request.id)
  }
}
