<?php

include 'logging.php';

function getImage()
{
    $limit = 20;
    $pageOffset = rand(0, 30000/$limit);
    $url = "https://gelbooru.com/index.php?page=dapi&s=post&q=index&limit=$limit&tags=loli+solo&pid=$pageOffset";
    logMessage("getting image from: $url\n");
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

    $possible = [];
    $doc = simplexml_load_string($result);
    foreach ($doc->post as $post)
    {
        // find the urls for each 'questionable' image
        if ($post["rating"] == "q")
        {
            array_push($possible, "https:" . (string) $post["file_url"]);
        }
    }
    if (count($possible) == 0)
    {
        return false;
    }

    return $possible[array_rand($possible)];
}
?>
