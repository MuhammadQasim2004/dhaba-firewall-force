<?php

require __DIR__ . '/../vendor/autoload.php';

use App\Http\Controllers\StatsController;

header('Content-Type: application/json');

$path = parse_url($_SERVER['REQUEST_URI'] ?? '/', PHP_URL_PATH);

if ($path === '/health') {
    echo json_encode(['status' => 'ok']);
    return;
}

if ($path === '/info') {
    echo json_encode([
        'service' => 'stats-api',
        'version' => getenv('GIT_COMMIT') ?: 'unknown',
    ]);
    return;
}

if ($path === '/stats') {
    $controller = new StatsController();
    echo json_encode($controller->stats());
    return;
}

http_response_code(404);
echo json_encode(['error' => 'not found']);
