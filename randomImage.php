<?php

include_once 'logging.php';

abstract class ImageType
{
    const baseUrl = "https://gelbooru.com/index.php?page=dapi&s=post&q=index&tags=";
    
    const SOLO = 1;
    const GROUP = 2;
}

function getImageForMessage($msg)
{
    $type = checkImageType($msg);
    if ($type === ImageType::SOLO)
    {
        return getSoloImage();
    }
    else if ($type === ImageType::GROUP)
    {
        return getMultipleImage();
    }
    logMessage("error determining image type for $msg", true);
    return false;
}

function checkImageType($msg)
{
    if (stripos($msg, 'me irl') !== false)
    {
        return ImageType::SOLO;
    }
    else if ((stripos($msg, 'me on the left') !== false) ||
             (stripos($msg, 'me on the right') !== false))
    {
        return ImageType::GROUP;
    }
    return false;
}

function getMatchingString($msg)
{
    $strings = ['me irl', 'me on the left', 'me on the right'];
    foreach($strings as $str)
    {
        if (stripos($msg, $str) !== false)
        {
            $startIndex = stripos($msg, $str);
            return substr($msg, $startIndex, strlen($str));
        }
    }
    logMessage("error determining string for $msg", true);
    return "";
}

function getMultipleImage()
{
    return getImage(3500, "score:>5 multiple_girls loli rating:questionable -large_breasts -1boy -multiple_boys");
}

function getSoloImage()
{
    return getImage(10000, "loli solo rating:questionable score:>5");
}

// TODO: get count from API instead of having callers hard code it
function getImage($max, $url)
{
    $limit = 1;
    $pageOffset = rand(0, $max/$limit);
    $url = ImageType::baseUrl . $url . "&limit=$limit&pid=$pageOffset";
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
    if (strlen((string) $doc->post[0]["file_url"]) > 0)
    {
        return "https:" . (string) $doc->post[0]["file_url"];
    }
    return false;
}
?>
