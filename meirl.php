<?php

include_once 'randomImage.php';
include_once 'logging.php';
include_once 'state.php';

// originally based on http://www.wikihow.com/Develop-an-IRC-Bot

// config parameters
$server = "ssl://irc.rizon.net";
$port = 6697;
$nickname = "meirlBot";
$ident = "meirl";
$gecos = "a bot to post pics of yourself irl";
$password = file_get_contents("password.txt");

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

$cooldown = [];
$joinedChannels = false;

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
        logMessage($data, false);
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
        continue; // nothing else to do with a ping
    }
    
    // identify after MOTD ends
    // :server 376 nick :End of /MOTD command.
    if ($d[1] === '376' || $d[1] === '422')
    {
        logMessage("identifying");
        fwrite($socket, "PRIVMSG nickserv :identify $password\r\n");
    }
    
    // join channels after vhost is set
    //:server 396 nick host :is now your visible host
    if ($d[1] == "396" && $d[2] == $nickname)
    {
        if ($joinedChannels == false)
        {
            $channels = State::getChannels();
            foreach ($channels as $channel)
            {
                if (strlen($channel) > 1)
                {
                    logMessage("joining $channel");
                    fwrite($socket, "JOIN $channel\r\n");
                }
            }
            echo "ready\n";
            $joinedChannels = true;
        }
    }
    
    // reply to messages
    // :nick!ident@host PRIVMSG #channel :message
    if ($d[1] == "PRIVMSG")
    {
        $msg = implode(' ', array_slice($d, 3));
        $nickEndIndex = strpos($d[0], "!");
        $otherNick = substr($d[0], 1, $nickEndIndex-1);
        if (checkImageType($msg) !== false)
        {
            // reply to the channel or to a pm
            $sendTo = false;
            if (stripos($d[2], '#') !== false && @$cooldown[$d[2]][$otherNick] == null)
            {
                $sendTo = $d[2];
                $cooldown[$d[2]][$otherNick] = 5;
                logMessage("cd for $otherNick is " . $cooldown[$d[2]][$otherNick]);
            }
            else if ($d[2] == $nickname)
            {
                $sendTo = $otherNick;
            }
            
            if ($sendTo !== false)
            {
                sendImage($socket, $sendTo, $msg);
                continue; // make sure we don't send multiple things due to one line
            }
        }
        else if (@$cooldown[$d[2]][$otherNick] !== null)
        {
            $cooldown[$d[2]][$otherNick] -= 1;
            logMessage("cd for $otherNick is " . $cooldown[$d[2]][$otherNick]);
            if ($cooldown[$d[2]][$otherNick] == 0)
            {
                $cooldown[$d[2]][$otherNick] = null;
            }
        }
    }
    
    // join from invites
    // :nick!ident@host INVITE nick :#channel
    if ($d[1] == "INVITE" && $d[2] == $nickname)
    {
        $channel = substr($d[3], 1);
        logMessage("joining $channel, invited by " . $d[0]);
        fwrite($socket, "JOIN $channel\r\n");
        // fwrite($socket, "PRIVMSG $channel :Me IRL Bot requested by " . substr($d[0], 1) . "\r\n");
        State::addChannel($channel);
        continue;
    }
    
    // we got kicked
    // :nick!ident@host KICK #channel nick :message
    if ($d[1] == "KICK" && $d[3] == $nickname)
    {
        logMessage("kicked from " . $d[2] . " by " . $d[0] . " because " . $d[4]);
        State::removeChannel($d[2]);
    }
    
    // we're banned from this channel
    // :server 474 nick channel :Cannot join channel (+b)
    if ($d[1] == "474" && $d[2] = $nickname)
    {
        logMessage("banned from " . $d[3]);
        State::removeChannel($d[3]);
    }
}

function sendImage($socket, $to, $msg)
{
    $url = getImageForMessage($msg);
    if ($url !== false)
    {
        $meirlString = getMatchingString($msg);
        $newMsg = "PRIVMSG " . $to . " " . ":$meirlString $url" . "\r\n";
        logMessage("sending image: ". $newMsg);
        fwrite($socket, $newMsg);
    }
}
?>
