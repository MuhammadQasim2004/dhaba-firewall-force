<?php

use App\Http\Controllers\StatsController;
use PHPUnit\Framework\TestCase;

final class StatsControllerTest extends TestCase
{
    public function testStatsShapeIsCorrectEvenWithoutRedis(): void
    {
        $controller = new StatsController(null);
        putenv('REDIS_HOST=invalid-host-for-test');
        $result = $controller->stats();

        $this->assertArrayHasKey('orders_per_minute', $result);
        $this->assertArrayHasKey('top_dishes', $result);
    }
}
