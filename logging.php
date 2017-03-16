<?php

function logMessage($msg, $alsoPrint=true)
{
    static $logFile = "log.txt";
    static $file = false;
    
    if (!is_resource($file))
    {
        echo "opening file $logFile\n";
        $file = fopen($logFile, 'a', true);
    }
    fwrite($file, $msg . "\n");
    
    if ($alsoPrint)
    {
        echo $msg . "\n";
    }
}

?>
