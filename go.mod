module github.com/unikiosk/unikiosk

go 1.17

require (
	filippo.io/mkcert v0.0.0-00010101000000-000000000000
	github.com/davecgh/go-spew v1.1.1
	github.com/elazarl/goproxy v0.0.0-20211114080932-d06c3be7c11b
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/joho/godotenv v1.4.0
	github.com/kbinani/screenshot v0.0.0-20210720154843-7d3a670d8329
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/mattn/go-gtk v0.0.0-20191030024613-af2e013261f5
	github.com/peterbourgon/diskv v2.0.1+incompatible
	github.com/phayes/freeport v0.0.0-20180830031419-95f893ade6f2
	github.com/spf13/cobra v1.2.1
	github.com/stretchr/testify v1.7.0
	github.com/zserge/lorca v0.1.10
	go.uber.org/zap v1.19.1
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/tools v0.1.5
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	honnef.co/go/tools v0.0.1-2020.1.6
)

require (
	github.com/BurntSushi/toml v0.3.1 // indirect
	github.com/felixge/httpsnoop v1.0.1 // indirect
	github.com/gen2brain/shm v0.0.0-20200228170931-49f9650110c5 // indirect
	github.com/google/btree v1.0.1 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jezek/xgb v0.0.0-20210312150743-0e0f116e1240 // indirect
	github.com/lxn/win v0.0.0-20210218163916-a377121e959e // indirect
	github.com/mattn/go-pointer v0.0.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/mod v0.5.1 // indirect
	golang.org/x/net v0.0.0-20211020060615-d418f374d309 // indirect
	golang.org/x/sys v0.0.0-20210615035016-665e8c7367d1 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	howett.net/plist v0.0.0-20181124034731-591f970eefbb // indirect
	software.sslmate.com/src/go-pkcs12 v0.0.0-20180114231543-2291e8f0f237 // indirect
)

replace (
	filippo.io/mkcert => github.com/FiloSottile/mkcert v0.0.0-20210213023452-0a3190b1659e
	github.com/zserge/lorca => github.com/unikiosk/lorca v0.0.0-20211120113058-786a2006febc
)
