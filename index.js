module.exports = (robot) => {
  // Get an express router to expose new HTTP endpoints
  const app = robot.route('/bot');

  // Add a new route
  app.get('/ping', (req, res) => {
    res.end('Pong!');
  });
  
  console.log('Yay, the app was loaded!')

  require('./scripts/assign-new-pr-to-review.js')(robot);
  require('./scripts/assign-to-bounty-awaiting-for-approval.js')(robot);
  require('./scripts/greet-new-contributor.js')(robot);

  // For more information on building apps:
  // https://probot.github.io/docs/

  // To get your app running against GitHub, see:
  // https://probot.github.io/docs/development/
}
