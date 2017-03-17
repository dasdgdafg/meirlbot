<?php

include_once 'logging.php';

class State {
    private static $channelFileName = "state_channels.txt";

    public static function addChannel($channel)
    {
        $contents = file_get_contents(self::$channelFileName);
        if (stripos($contents, $channel . "\n") !== false)
        {
            logMessage("already added $channel");
        }
        else
        {
            logMessage("adding $channel");
            file_put_contents(self::$channelFileName, $contents . $channel . "\n");
        }
    }
    
    public static function removeChannel($channel)
    {
        $contents = file_get_contents(self::$channelFileName);
        if (stripos($contents, $channel . "\n") === false || strlen($channel) < 2)
        {
            logMessage("don't know about $channel\n");
        }
        else
        {
            logMessage("removing $channel");
            $contents = str_replace($channel . "\n", '', $contents);
            file_put_contents(self::$channelFileName, $contents);
        }
    }
    
    public static function getChannels()
    {
        $contents = file_get_contents(self::$channelFileName);
        return explode("\n", $contents);
    }
}

?>
