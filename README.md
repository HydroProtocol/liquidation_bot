# DDEX Liquidation Bot

Participate in the liquidation on ddex to earn money 💰💰💰

## Getting Started

Margin position will be liquidated if the collateral rate too low. 
When a liquidation occurs, the borrowers' assets are sold off in a form of a dutch auction to repay the loan. 
This bot help you bid the auction when profitable.

The bot uses an arbitrage strategy: Sell collateral on DDEX immediately after receive it from an auction. For example:

 - An auction sell *10ETH* at price *160USDT*
 - The bot buy *10ETH*
 - The bot sell *10ETH* on ddex at price *170USDT*
 - Earn *100USDT*

### Prerequisites

In order to run this container you'll need docker installed.

* [Windows](https://docs.docker.com/windows/started)
* [OS X](https://docs.docker.com/mac/started/)
* [Linux](https://docs.docker.com/linux/started/)

You need to prepare some asset(e.g. ETH, USDT, DAI) in your DDEX spot trading balance. 

### Run Container

```shell
docker run -it -v /your/file/path:/workingDir --name=auctionBidder hydroprotocolio/liquidation_bot:latest /bin/main
```

If you want to test on ropsten

```
docker run -it -v /your/file/path:/workingDir --name=auctionBidder --env NETWORK=ropsten hydroprotocolio/liquidation_bot:latest /bin/main
```

#### Volumes

* `/your/file/path` - Where liquidation history, config and logs stored

#### Parameters

The bot will ask for the following parameters for the first time and stored them at `/your/file/path/config.json`. Edit this file to adjust bot parameters.

* `PRIVATE_KEY` - Private key of the account to join liquidation

* `ETHEREUM_NODE_URL` - Ethereum node url. Get a free node at [infura](https://infura.io).

* `MARKETS` - Which markets' auction I am interested in. Separated by commas `ETH-USDT,ETH-DAI` 
	
* `MIN_AMOUNT_VALUE_USD` - I don't want to participate the auction unless its USD value greater than MIN_AMOUNT_VALUE_USD

* `PROFIT_BUFFER` - I don't want to bid unless the 
	
* `MAX_SLIPPAGE` - 
	

	
* `GAS_PRICE_TIPS_IN_GWEI` -

## Running the tests

Explain how to run the automated tests for this system

### Break down into end to end tests

Explain what these tests test and why

```
Give an example
```

### And coding style tests

Explain what these tests test and why

```
Give an example
```

## Deployment

Add additional notes about how to deploy this on a live system

## Built With

* [Dropwizard](http://www.dropwizard.io/1.0.2/docs/) - The web framework used
* [Maven](https://maven.apache.org/) - Dependency Management
* [ROME](https://rometools.github.io/rome/) - Used to generate RSS Feeds

## Contributing

Please read [CONTRIBUTING.md](https://gist.github.com/PurpleBooth/b24679402957c63ec426) for details on our code of conduct, and the process for submitting pull requests to us.

## Versioning

We use [SemVer](http://semver.org/) for versioning. For the versions available, see the [tags on this repository](https://github.com/your/project/tags). 

## Authors

* **Billie Thompson** - *Initial work* - [PurpleBooth](https://github.com/PurpleBooth)

See also the list of [contributors](https://github.com/your/project/contributors) who participated in this project.

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details

## Acknowledgments

* Hat tip to anyone whose code was used
* Inspiration
* etc
