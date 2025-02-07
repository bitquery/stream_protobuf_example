# stream_protobuf_example
Example of project reading protobuf stream from kafka

## Run project

checkout:

```
git clone https://github.com/bitquery/stream_protobuf_example.git
```

install golang (https://go.dev/doc/install) and compile:

you may need to install C++

```
apt-get install g++
```

and then 

```
cd stream_protobuf_example/
go build
```



copy config:

```
cp config_example.yml config.yml
```

then fill in variables ```<YOUR USERNAME>``` and ```<YOUR PASSWORD>``` in config.yml file


You also can change the desired topic name:

Protobuf format, no block aggregation
```
solana.dextrades.proto
solana.tokens.proto
solana.transactions.proto
```

and json format, aggregated by block
```
solana.balance_updates
solana.dexorders
solana.dexpools
solana.dextrades
solana.instruction_balance_updates
solana.instructions
solana.transfers
```

run it:

```
./stream_protobuf_example 
%4|1738918885.417|CONFWARN|rdkafka#producer-1| [thrd:app]: Configuration property group.id is a consumer property and will be ignored by this producer instance
%4|1738918885.417|CONFWARN|rdkafka#producer-1| [thrd:app]: Configuration property enable.auto.commit is a consumer property and will be ignored by this producer instance
Subscribing to topic solana.dextrades.proto 4 partitions: [{0 Success 3 [3] [3]} {1 Success 2 [2] [2]} {2 Success 2 [2] [2]} {3 Success 1 [1] [1]}]...
Assigned 4 consumers to solana.dextrades.proto topic
press Ctrl-C to exit
Running consumer rdkafka#consumer-2
Running consumer rdkafka#consumer-5
Running consumer rdkafka#consumer-3
Running consumer rdkafka#consumer-4
slot 319062398 processed 2 txs (2 trades) from partition 3[894333] in worker 3
slot 319062398 processed 1 txs (1 trades) from partition 3[894336] in worker 7
slot 319062398 processed 4 txs (6 trades) from partition 3[894332] in worker 4
slot 319062398 processed 5 txs (8 trades) from partition 3[894335] in worker 5
...
```

after pressing Ctrl-C it will show basic statistical report

## Checking speed of message reading

To test how fast your server reads messages, try the following experiment:

1. Set ```enable.auto.commit: true``` in config.yml and set ```group.id: <YOUR USERNAME>_<SOME RANDOM NUMBER>```
2. Run consumer for some time, stop it. Wait a minute and start again

What you expect to see is the reducing lag in messages, from approx 60,000 msec to close to 0.

It must catch up in less than 10 seconds. If it does not catch up fast, test server connection speed and ping to brokers.
