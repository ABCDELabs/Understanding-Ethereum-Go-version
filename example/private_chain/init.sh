# 1. First init new genesis state
geth init --datadir  <Datadir> genesis.json

# 2. Then, start geth Client.
geth --datadir  <Datadir>  --datadir <datadir> --http --http.corsdomain https://remix.ethereum.org --http.api personal,eth,net,web3,debug --syncmode=full console 2

