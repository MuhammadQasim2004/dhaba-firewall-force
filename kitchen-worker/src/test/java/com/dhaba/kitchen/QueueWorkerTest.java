package com.dhaba.kitchen;

import org.junit.jupiter.api.Test;
import org.springframework.jdbc.core.JdbcTemplate;
import redis.clients.jedis.JedisPool;

import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;

/**
 * Unit tests for the order state transition logic. We mock JdbcTemplate
 * and JedisPool so these tests run with no real Postgres or Redis —
 * that's the responsibility of integration tests, not this suite.
 */
class QueueWorkerTest {

    @Test
    void processOrderMovesStatusFromCookingToReady() {
        JdbcTemplate jdbcTemplate = mock(JdbcTemplate.class);
        JedisPool jedisPool = mock(JedisPool.class);
        QueueWorker worker = new QueueWorker(jedisPool, jdbcTemplate);

        worker.processOrder(42L);

        // first transition: new -> cooking
        verify(jdbcTemplate).update(
                eq("UPDATE orders SET status = 'cooking' WHERE id = ?"), eq(42L));
        // second transition: cooking -> ready
        verify(jdbcTemplate).update(
                eq("UPDATE orders SET status = 'ready' WHERE id = ?"), eq(42L));
    }
}
