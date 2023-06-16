## Run a private network
Run a private network of 3 validator nodes

1. Generate node keys:
```shell
../build/devp2p key generate node[1..3].key
```

2. Create 3 validators:
```shell
../build/sether validator new
```
V1:
- Public key: 0xc004135f0d5860bc30341d22cc44f3007d0bf35ee815cc827215c96d7d9aba0fb906c5e8c6eb651cf24625b9151cb2e919259b4f006301292d158e5415d66564b81e
- Address:    0xf9215294250dF0D4Beb912C8e18F1e1416d6A398
- Private key: 794ab11e60f5558a04680627eeebf6eb57201ff4e7f686926cc960c75f781ee2

V2:
- Public key: 0xc00459be00a14b8bd3b249ab7914e44f5c8e01be92aeafc51ec07f523c965cd74bc821cd7389c2ee9d87aaea327c13b6e4027c9cfcc651ef9b52a2adaeb69e6d6142
- Address:    0x594f344E2B6662C4b6c5A07B6b1287c6209c9c22
- Private key: e71c7ec6474d099b5b610919e4895370a6f4d52e02ab73e764701b86b4133040

V3:
- Public key: 0xc0047a6230b289af747663ccff2c95edd2061b029b4e888847b2b8aed005e22daafe9a39f1bb5f12044adc25a3e8e733bb7b62088b746eec8ed2fee9e131ece5e907
- Address:    0x280620317a56474ABCcc05d7Af612C8D11956611
- Private key: 1521b5eec58809def56686342e86075b1aad3731bd1a3ba01dc5bcb41f687104

3. Create 3 accounts:
```shell
../build/sether account new

Accounts `ID: Public key (Private key)`
```
- A1: 0x01D4d20f19315D78f5E942029345dad1e85fce55 (8ba12e09a11cbd6a97c4c60032ebe1e5c69dde7c272248a5ba35c5fd91bcd450) 
- A2: 0x4a14c36f2A8D73525D44E70Fa2EaA2483A916690 (bad5ce385e0226f8f454c6dc82f3edecaf9b1d802d60de1ad6ae1654be9d70d1)
- A3: 0x349e543718458B46244f958e7BA4a5c2848F9c78 (0ab73e1c4b6aed9f363d39c6a059896601defbd7f0a1357188d2763a471f3e2d)

4. Update Testnet info in `cmd/sether/launcher/creategen.go`

5. Build the new binary: `make sether`

