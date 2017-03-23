<?php

include_once 'logging.php';

function getImage()
{
    $limit = 1;
    $pageOffset = rand(0, 10000/$limit);
    $url = "https://gelbooru.com/index.php?page=dapi&s=post&q=index&limit=$limit&tags=loli+solo+rating:questionable+score:>5&pid=$pageOffset";
    logMessage("getting image from: $url");
    $options = array(
        'http' => array(
            'method'  => 'GET',
        )
    );
    $context  = stream_context_create($options);
    $result = file_get_contents($url, false, $context);
    if ($result === FALSE) 
    { 
        return false;
    }

    $doc = simplexml_load_string($result);
    return "https:" . (string) $doc->post[0]["file_url"];
}
?>
