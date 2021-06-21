const hre       = require("hardhat");
const pairsData = require("./pairsData.json");
const abi       = require("./artifacts/contracts/BurnCalc.sol/BurnCalc.json");
const fs        = require("fs");
const ethers    = hre.ethers;
const WETH      = "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2".toLowerCase();

async function main() {
    await hre.network.provider.request({
        method: "hardhat_reset",
        params: [{
            forking: {
                jsonRpcUrl: "EHTER YOUR RPCURL",
                blockNumber: 12588860
            }
        }]
    });

    const accounts = await ethers.getSigners();

    const signer = accounts[0];

    const balance = await ethers.provider.getBalance(accounts[0].address);
    console.log(balance);

    await ethers.provider.getBlockNumber().then((blockNumber) => {
      console.log("Current block number: " + blockNumber);
    });

    var ass = await hre.run("compile");

    const Burn = await hre.ethers.getContractFactory("BurnCalc");
    const burn = await Burn.deploy({gasLimit: 4245000});

    await burn.deployed();

    const contractWithSigner = burn.connect(signer);

    var gas = await signer.getGasPrice();
    var add = ethers.BigNumber.from(200)
    const sentETH = await signer.sendTransaction({to: burn.address, value: ethers.utils.parseEther("7000")});
    const confirmedETH = await sentETH.wait();
    var noncer = await signer.getTransactionCount();
    const receipt2 = await contractWithSigner.convertETHtoWETH(WETH, ethers.utils.parseEther("6999"),{gasPrice: gas.add(add), gasLimit: 3000000, nonce: noncer });
    const confirmed = await receipt2.wait();
    console.log(pairsData["data"].length)

//    const contractBalance = await ethers.provider.getBalance(burn.address);
//    const wethBalance  = await burn.getWETHBalance(WETH);
    var count = 0;
    fs.writeFile('/path/to/tokenFees.json',`{ "data": [`, { flag: 'a+' }, err => console.log(err));
    for (var pair of pairsData["data"]) {
        var noncer = await signer.getTransactionCount()
        const fromIsToken0 = WETH == pair.token0;
        const amount = ethers.utils.parseEther("0.1");
        var rec;
        try {
            rec = await contractWithSigner.V2Buy(fromIsToken0, pair.pairAddress, amount, {gasPrice: ethers.BigNumber.from(800000000000), gasLimit: 800000, nonce: noncer});
        } catch (err) {
            console.log(`ERROR: ${err}`);
            continue;
        }
        const confirmed = await rec.wait();
        if (fromIsToken0) {
            const burnt = await burn.getBurnFactor(pair.token1);
            if (burnt[0].eq(burnt[1]) ) {
                fs.writeFile('/path/to/tokenFees.json',`{"tokenAddress": "${pair.token1}", "fee": 0},`, { flag: 'a' }, err => console.log(err));
            } else {
                const diff = burnt[0].sub(burnt[1]);
                const burnFactor = (diff.mul(ethers.BigNumber.from(10000)).div(burnt[0])).toNumber();
                const burnFactorPercentage = burnFactor/100.0;
                fs.writeFile('/path/to/tokenFees.json',`{"tokenAddress": "${pair.token1}", "fee": ${burnFactorPercentage}},`, { flag: 'a' }, err => console.log(err));
            }
        } else {
            const burnt = await burn.getBurnFactor(pair.token0);
            if (burnt[0].eq(burnt[1])) {
                fs.writeFile('/path/to/tokenFees.json',`{"tokenAddress": "${pair.token0}", "fee": 0},`, { flag: 'a' }, err => console.log(err));
            } else {
                const diff = burnt[0].sub(burnt[1]);
                const burnFactor = (diff.mul(ethers.BigNumber.from(10000)).div(burnt[0])).toNumber();
                const burnFactorPercentage = burnFactor/100.0;
                fs.writeFile('/path/to/tokenFees.json',`{"tokenAddress": "${pair.token0}", "fee": ${burnFactorPercentage}},`, { flag: 'a' }, err => console.log(err));
            }
        }
        count++;
        console.log(`Proccessed ${count} transactions`);
        // So we don't go over gas block limit
        if (count % 20 === 0) {
            await hre.network.provider.request({
                method: "hardhat_reset",
                params: [{
                    forking: {
                        jsonRpcUrl: "YOUR RPC URL",
                        blockNumber: 12588860
                    }
                }]
            });
            var ass = await hre.run("compile");

            const Burn = await hre.ethers.getContractFactory("BurnCalc");
            const burn = await Burn.deploy({gasLimit: 4245000});

            await burn.deployed();

            const contractWithSigner = burn.connect(signer);

            var gas = await signer.getGasPrice();
            var add = ethers.BigNumber.from(200)
            const sentETH = await signer.sendTransaction({to: burn.address, value: ethers.utils.parseEther("9000")});
            const confirmedETH = await sentETH.wait();
            var noncer = await signer.getTransactionCount();
            const receipt2 = await contractWithSigner.convertETHtoWETH(WETH, ethers.utils.parseEther("8000"),{gasPrice: ethers.BigNumber.from(800000000000), gasLimit: 800000, nonce: noncer });
            const confirmed = await receipt2.wait();

        }
    }
    fs.writeFile('path/to/tokenFees.json',`]}`, { flag: 'a+' }, err => console.log(err));
}

main()
