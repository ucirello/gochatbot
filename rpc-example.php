<?php
// This implementation executes a echo service in gochatbot and PHP.
$rpcServer = getenv('GOCHATBOT_RPC_BIND');

function botPop($rpcServer) {
	$raw = file_get_contents(sprintf('http://%s/pop', $rpcServer));

	$ret = json_decode($raw, true);
	$err = json_last_error();
	if (JSON_ERROR_NONE != $err) {
		die(json_last_error_msg());
	}

	return $ret;
}

function botSend($rpcServer, $msg) {
	$url = sprintf('http://%s/send', $rpcServer);

	$json = json_encode($msg);
	$err = json_last_error();
	if (JSON_ERROR_NONE != $err) {
		die(json_last_error_msg());
	}

	$options = [
		'http' => [
			'header' => "Content-type: application/json\r\n",
			'method' => 'POST',
			'content' => $json,
		],
	];
	$context = stream_context_create($options);
	return file_get_contents($url, false, $context);
}

while (true) {
	$msg = botPop($rpcServer);
	if (empty($msg['Message'])) {
		continue;
	}
	echo 'Got:', PHP_EOL;
	print_r($msg);

	$newMsg = [
		'Room' => $msg['Room'],
		'FromUserID' => $msg['ToUserID'],
		'FromUserName' => $msg['ToUserName'],
		'ToUserID' => $msg['FromUserID'],
		'ToUserName' => $msg['FromUserName'],
		'Message' => 'echo: ' . $msg['Message'],
	];

	echo 'Sending:', PHP_EOL;
	print_r($newMsg);

	botSend($rpcServer, $newMsg);
}