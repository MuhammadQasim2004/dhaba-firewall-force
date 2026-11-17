<?php

namespace App\Http\Controllers;

use Predis\Client as Redis;

class StatsController
{
    private Redis $redis;

    public function __construct(?Redis $redis = null)
    {
        $host = getenv('REDIS_HOST') ?: 'redis';
        $this->redis = $redis ?? new Redis(['host' => $host]);
    }

    /** @return array{orders_per_minute: int, top_dishes: string[]} */
    public function stats(): array
    {
        $ordersPerMinute = 0;
        $topDishes = [];

        try {
            $ordersPerMinute = (int) ($this->redis->get('stats:orders_per_minute') ?? 0);
            $topDishes = $this->redis->zrevrange('stats:top_dishes', 0, 4) ?: [];
        } catch (\Throwable $e) {
            // Redis unreachable: return zeros rather than a 500 so the
            // stats panel degrades gracefully.
        }

        return [
            'orders_per_minute' => $ordersPerMinute,
            'top_dishes' => $topDishes,
        ];
    }
}
