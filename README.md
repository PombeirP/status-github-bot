# status-github-bot
A bot for github

## Creating the bot GitHub App
This bot is packaged as a GitHub App. To install it, one needs to follow the procedure to create a GitHub App (only needs to be done once and can be made public for any number of repositories).
1. Create the GitHub App:
    1. In GitHub, go to `Settings/Developer settings/GitHub Apps` and click on `New GitHub App`
    1. Enter the bot name in `GitHub App name`, e.g. `Status GitHub Bot`
    1. In `Homepage URL`, enter the `/bot.log` endpoint of the service, e.g. https://5e63b0ab.ngrok.io/bot.log
    1. In `Webhook URL`, enter the `/webhook` endpoint of the service, e.g. https://5e63b0ab.ngrok.io/webhook
    1. In `Webhook secret (optional)`, enter a string of characters that matches the value in the config.json file deployed with the service.
    1. The app needs `Read-only` permission to `Pull requests`, `Read & write` permission to `Repository projects`.
    1. The app subscribes to `Pull request` events.
    1. Generate a private key, and save it in the root folder of the service, with the following file name: 
`status-github-bot.private-key.pem`
1. Installing the bot service:
    1. Compile the bot and deploy it to the cloud.
    1. Make sure the `config.json` and `status-github-bot.private-key.pem` files are present in the same directory as the executable.
1. Install the GitHub App in an account:
    1. Select the repositories where the bot should work (e.g. status-react)
