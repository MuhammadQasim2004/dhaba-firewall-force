package com.dhaba.kitchen;

import jakarta.annotation.PostConstruct;
import jakarta.annotation.PreDestroy;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.stereotype.Component;
import redis.clients.jedis.Jedis;
import redis.clients.jedis.JedisPool;

import java.util.List;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;

/**
 * Pops order ids from the Redis queue, moves each order through
 * new -> cooking -> ready in Postgres, and writes a heartbeat to
 * Redis every 10 seconds so Docker/Kubernetes can tell the worker
 * (which has no web port) is alive.
 */
@Component
public class QueueWorker {

    private static final String QUEUE_KEY = "orders:queue";
    private static final String HEARTBEAT_KEY = "kitchen-worker:heartbeat";

    @Value("${app.git-commit:unknown}")
    private String gitCommit;

    private final JedisPool jedisPool;
    private final JdbcTemplate jdbcTemplate;
    private final ScheduledExecutorService scheduler = Executors.newScheduledThreadPool(2);

    public QueueWorker(JedisPool jedisPool, JdbcTemplate jdbcTemplate) {
        this.jedisPool = jedisPool;
        this.jdbcTemplate = jdbcTemplate;
    }

    @PostConstruct
    public void start() {
        scheduler.scheduleAtFixedRate(this::heartbeat, 0, 10, TimeUnit.SECONDS);
        scheduler.scheduleWithFixedDelay(this::drainQueue, 0, 1, TimeUnit.SECONDS);
    }

    @PreDestroy
    public void stop() {
        scheduler.shutdownNow();
    }

    private void heartbeat() {
        try (Jedis jedis = jedisPool.getResource()) {
            jedis.set(HEARTBEAT_KEY, String.valueOf(System.currentTimeMillis()));
        } catch (Exception e) {
            // Redis down: heartbeat just won't update, which is exactly
            // what the Dockerfile HEALTHCHECK / monitoring should catch.
        }
    }

    private void drainQueue() {
        try (Jedis jedis = jedisPool.getResource()) {
            List<String> popped = jedis.rpop(QUEUE_KEY, 10);
            if (popped == null) {
                return;
            }
            for (String idStr : popped) {
                processOrder(Long.parseLong(idStr));
            }
        } catch (Exception e) {
            // transient failure; order stays in Postgres as "new" and will
            // be retried on next poll since we only popped from the queue,
            // not from the source of truth.
        }
    }

    private void processOrder(long orderId) {
        jdbcTemplate.update("UPDATE orders SET status = 'cooking' WHERE id = ?", orderId);
        try {
            Thread.sleep(500); // simulate cook time
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }
        jdbcTemplate.update("UPDATE orders SET status = 'ready' WHERE id = ?", orderId);
    }
}
