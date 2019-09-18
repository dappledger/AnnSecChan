# AnnSecChan

AnnSecChan is the server witch provide private transactions. This server should running with AnnChain.

## Requirements

| Requirement | Notes              |
| ----------- | ------------------ |
| Go version  | Go1.12.0 or higher |

## Building the source 

git clone https://github.com/dappledger/AnnSecChan.git

cd AnnSecChan/cmd

go build

## Configuration 

{

	"port":"8000",				### service listen net port.
	
	"log_dir":"./log",			### logs storage directory
	
	"db_path":"./data",         ### data storage directory
	
	"log_level":0,			    ### logs print level: Debug=0 Info=1 Warn=2	Error=3 Fatal=4 Read=5 Update=6
	
	"p2p_moniker":"node_0",	    ### node nickname
	
	"p2p_privkey":"",           ### node ed25519 privkey,the same as chain node privkey.
	
	"p2p_listen_addr":"tcp://127.0.0.1:26658",    ### p2p listen net address
	
	"p2p_peers":"192.168.0.1:26657,192.168.0.2:26657",	### peers p2p net address,if multiple separated with comma
	
	"p2p_blacklist_pubkey":"",  ### refuse pubkey list,if multiple pubkeys separated with comma
	
	"p2p_whitelist_pubkey":""   ### only access pubkey list,if multiple pubkeys separated with comma 
	
}


## Quick Start

./cmd -c config.json


## RESTFUL API

###/v1/transaction
Method: PUT
Request:
{
	"public_keys":["971D0EB6F0FECA0B7365E621FD9EC5E6D281604DBDD82A3A85931F62B19AE7F9"], 
	"value": "MTIzNDU2Nzg="  
}

Response:
{
  "data": "",
  "isSuccess": true,
  "message": "Success"
}


###/v1/transaction/:key 
Method: GET
Request:
private data hash,such as "/v1/transaction/0x8696933513c80d6d8d5c7ecea31740c659824a6090ddad2d5d575def0669daec"
Response:
{
  "data": "MTIzNDU2Nzg=",
  "isSuccess": true,
  "message": ""
}

###/v1/node/peers    
Method: GET
Response:
{
  "data": [
    {
      "moniker": "node_2",
      "address": "172.17.32.26:26658",
      "pubkey": "971D0EB6F0FECA0B7365E621FD9EC5E6D281604DBDD82A3A85931F62B19AE7F9"
    }
  ],
  "isSuccess": true,
  "message": ""
}
