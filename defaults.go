package main

const _conf_default = `
[movie-service]
uri = 127.0.0.1:12345

[rabbitmq]
uri = amqp://guest:guest@localhost:5672/    ; AMQP URI amqp://user:pass@host:port
reliable = true                             ; Wait for the publisher confirmation before exiting

[rottentomatoes]
rottentomatoes_api_key = ; use your own key

`
const _log_default = `
<seelog minlevel="trace" maxlevel="critical">
    <!-- exceptions>
        <exception filepattern="test*" minlevel="error"/>
    </exceptions -->

    <outputs formatid="file">
        <filter levels="trace,debug,info,warn,error,critical" formatid="file">
            <file path="movie-service.log" />
        </filter>
        <filter levels="info,warn,error,critical" formatid="console">
            <console/>
        </filter>
    </outputs>

    <formats>
        <format id="file" format="%Date(2006-01-02 15:04:05.000) %FullPath:%Line [%LEV] %Msg%n"/>
        <format id="console" format="%Date(2006-01-02 15:04:05.000) [%LEV] %Msg%n"/>
    </formats>
</seelog>`
