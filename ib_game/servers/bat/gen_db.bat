set GOPATH=%cd%/../../../../../../..

cd ../conf/db
md proto

cd ../../tools
code_generator.exe -c ../conf/db/game_db.json -d ../game -p ../conf/db/proto/game_db.proto
code_generator.exe -c ../conf/db/account_db.json -d ../account -p ../conf/db/proto/account_db.proto

cd protobuf
protoc.exe --go_out=../../game/game_db --proto_path=../../conf/db/proto game_db.proto
protoc.exe --go_out=../../account/account_db --proto_path=../../conf/db/proto account_db.proto
cd ../../proxy/test
