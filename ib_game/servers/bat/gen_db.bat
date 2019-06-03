set GOPATH=%cd%/../../../../../../

cd db_define
md proto
cd ..

cd ../../bin
code_generator.exe -c ../proxy/test/db_define/game_db.json -d ../proxy/test -p ../proxy/test/db_define/proto/game_db.proto
cd ../example

cd protobuf
protoc.exe --go_out=../../proxy/test/game_db --proto_path=../../proxy/test/db_define/proto game_db.proto
cd ../../proxy/test

go build -i -o ../bin/test_client.exe github.com/huoshan017/mysql-go/proxy/test
if errorlevel 1 goto exit

goto ok

:exit
echo build failed !!!

:ok
echo build ok