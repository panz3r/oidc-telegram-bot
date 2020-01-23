# OIDC Telegram Bot
> A sample serverless Telegram bot with OIDC Authentication support built in Go

This sample consists of 2 AWS lambdas:
- Auth, lambda that handles OIDC post-auth redirection
- Telegram, lambda that handles the main Telegram bot conversation flow

## Usage

This project is built using the [`Serverless`](https://serverless.com) framework.

Clone repo

```bash
git clone https://github.com/panz3r/oidc-telegram-bot

cd oidc-telegram-bot
```

## Configure

- Copy `.env.sample.yml` to `.env.<stage>.yml` (based on deployment `stage`, e.g. `.env.dev.yml`)

- Fill in required info

### Build

To compile the 2 lambdas run 

```bash
make build
```

### Deploy 

To deploy the 2 lambdas to AWS run

```bash
make deploy
```

### Setup Telegram

Once the 2 lambdas are deployed, create a Telegram bot following the [official guide](https://core.telegram.org/bots) and then set the deployed Telegram AWS lambda URL as the bot webhook.

## Contributing
Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License
[MIT](https://choosealicense.com/licenses/mit/)