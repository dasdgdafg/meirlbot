<?php

include 'randomImage.php';

// based on http://www.wikihow.com/Develop-an-IRC-Bot

// config parameters
$server = "ssl://irc.rizon.net";
$port = 6697;
$nickname = "meirlBot";
$ident = "meirl";
$gecos = "is this used for anything?";
// $channel = "#sadpanda";
$channel = "#testing751984351";

// connect to the network
$socket = stream_socket_client("$server:$port");
if ($socket === false)
{
    $errorCode = socket_last_error();
    $errorString = socket_strerror($errorCode);
    die("Error $errorCode: $errorString\n");
}

// close the socket if we die
register_shutdown_function('shutdown');
function shutdown()
{
    fclose($socket);
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
        fwrite($socket, "PONG " . $d[1] . "\r\n");
    }
    
    // join the channel after MOTD ends
    if ($d[1] === '376' || $d[1] === '422')
    {
        fwrite($socket, "JOIN $channel\r\n");
    }
    
    // do stuff
    // :nick!ident@host PRIVMSG #channel :message
    if ($d[1] == "PRIVMSG" && stripos($d[2], '#') !== false)
    {
        $msg = implode(' ', array_slice($d, 3));
        if (stripos($msg, 'me irl') !== false)
        {
            $url = getImage();
            if ($url !== false)
            {
                $startIndex = stripos($msg, 'me irl');
                $meirlString = substr($msg, $startIndex, 6);
                $newMsg = "PRIVMSG " . $d[2] . " " . ":$meirlString $url" . "\r\n";
                echo "sending message: ". $newMsg;
                fwrite($socket, $newMsg);
            }
        }
    }
    else if ($d[1] == "PRIVMSG" && $d[2] == $nickname)
    {
        $msg = implode(' ', array_slice($d, 3));
        $nickEndIndex = strpos($d[0], "!");
        $otherNick = substr($d[0], 1, $nickEndIndex-1);
        if (stripos($msg, 'me irl') !== false)
        {
            $url = getImage();
            if ($url !== false)
            {
                $startIndex = stripos($msg, 'me irl');
                $meirlString = substr($msg, $startIndex, 6);
                $newMsg = "PRIVMSG " . $otherNick . " " . ":$meirlString $url" . "\r\n";
                echo "sending message: ". $newMsg;
                fwrite($socket, $newMsg);
            }
        }
    }
}

?>
