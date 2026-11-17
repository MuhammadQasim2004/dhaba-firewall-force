package com.dhaba.kitchen;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import redis.clients.jedis.JedisPool;

@Configuration
public class JedisConfig {

    @Value("${app.redis-host:redis}")
    private String redisHost;

    @Value("${app.redis-port:6379}")
    private int redisPort;

    @Bean
    public JedisPool jedisPool() {
        return new JedisPool(redisHost, redisPort);
    }
}
