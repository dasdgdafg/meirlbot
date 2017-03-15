<?php

// based on http://www.wikihow.com/Develop-an-IRC-Bot

// config parameters
$server = "ssl://irc.rizon.net";
$port = 6697;
$nickname = "testBot86545";
$ident = "testbot";
$gecos = "is this used for anything?";
$channel = "#testing751984351";

// connect to the network
// $socket = socket_create(AF_INET, SOCK_STREAM, SOL_TCP);

$socket = stream_socket_client("$server:$port");
if ($socket === false)
{
    $errorCode = socket_last_error();
    $errorString = socket_strerror($errorCode);
    die("Error $errorCode: $errorString\n");
}

// send the registration info
echo fwrite($socket, "NICK $nickname\r\n") . "\n";
echo fwrite($socket, "USER $ident * 8 :$gecos\r\n") . "\n";

// loop until the socket closes
while (is_resource($socket))
{
    // fetch data from the socket
    $data = fread($socket, 1024);
    if ($data === false)
    {
        $errorCode = socket_last_error();
        $errorString = socket_strerror($errorCode);
        die("Error $errorCode: $errorString\n");
    }
    $data = trim($data);
    if (strlen($data) > 0)
    {
        echo $data . "\n";
    }
    if (feof($socket))
    {
        die("socket is at eof\n");
    }
    
    // split into words
    $d = explode(' ', $data);
    
    // pad the array instead of checking the length later
    $d = array_pad($d, 10, "");
    
    // handle pings
    if ($d[0] === "PING")
    {
        echo "replying to ping with ". "PONG " . $d[1] . "\r\n";
        fwrite($socket, "PONG " . $d[1] . "\r\n");
    }
    
    // join the channel after MOTD ends
    if ($d[1] === '376' || $d[1] === '422')
    {
        echo "joining channel with ". "JOIN $channel\r\n";
        fwrite($socket, "JOIN $channel\r\n");
    }
    
    // do stuff
    // :nick!ident@host PRIVMSG #channel :message
    if ($d[2] == $channel)
    {
        $msg = implode(' ', array_slice($d, 3));
        if (strpos($msg, 'test') !== false)
        {
            echo "sending message with ". "PRIVMSG " . $d[2] . " " . str_replace('test', 'penis', $msg) . "\r\n";
            fwrite($socket, "PRIVMSG " . $d[2] . " " . str_replace('test', 'penis', $msg) . "\r\n");
        }
    }
}

?>
