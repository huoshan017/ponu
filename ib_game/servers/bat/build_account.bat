set GOPATH=%cd%/../../../../../../..

go build -i -o ../bin/account.exe github.com/huoshan017/ponu/ib_game/servers/account
if errorlevel 1 goto exit

goto ok

:exit
echo build failed !!!

:ok
echo build ok