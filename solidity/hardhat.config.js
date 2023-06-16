require("@nomicfoundation/hardhat-toolbox");
require('hardhat-abi-exporter');

/** @type import('hardhat/config').HardhatUserConfig */
module.exports = {
    gasReporter: {
        currency: 'USD',
        enabled: false,
        gasPrice: 50
    },
    contractSizer: {
        runOnCompile: true
    },
    abiExporter: {
        runOnCompile: false,
        path: './abi',
        clear: true,
        flat: true,
        spacing: 2
    },
    solidity: {
        version: "0.8.17",
        settings: {
            optimizer: {
                enabled: true,
                runs: 200
            }
        }
    },
    networks: {
        hardhat: {
            allowUnlimitedContractSize: true,
            chainId: 20244
        },
        fakenet: {
            url: 'http://localhost:18545',
            chainId: 20244,
            accounts: [
                "163f5f0f9a621d72fedd85ffca3d08d131ab4e812181e0d30ffd1c885d20aac7"
            ]
        },
        testnet: {
            url: 'https://rpc-test.sether.com',
            chainId: 20243,
            accounts: [
                "430306461c5326ef4d993cd91f205883d50b4c97e7b326b4018a1a383adfac59"
            ]
        }
    }
};
