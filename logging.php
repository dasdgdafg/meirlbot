<?php

function logMessage($msg, $alsoPrint=true)
{
    static $file = false;
    
    if (!is_resource($file))
    {
        $dateName = date(DATE_ATOM);
        $logFile = "logs/log-$dateName.txt";
        $logFile = str_replace(":", "-", $logFile);
        echo "opening file $logFile\n";
        $file = fopen($logFile, 'w', true);
    }
    fwrite($file, date(DATE_ATOM) . "\t" . $msg . "\n");
    
    if ($alsoPrint)
    {
        echo $msg . "\n";
    }
}
?>
