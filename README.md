# nats-proxy
Write to the [NATS](https://docs.nats.io/developing-with-nats/developer) via Memcached protocol.

# docker
See [Docker](https://hub.docker.com/r/raoptimus/memcached-nats-proxy)

## Why?

Because some script languages don't have a high performance NATS clients, but have a good Memcached client :)

### Example PHP client with connection pool.

```php
<?php
$mem = new Memcached('nats-streaming-connection-pool');
if (count($mem->getServerList()) == 0) {
    $mem->addServer("10.112.179.191", 11211);
    $mem->addServer("10.112.179.192", 11211);
    // http://php.net/manual/en/memcached.constants.php
    $mem->setOption(Memcached::OPT_TCP_NODELAY, true);  // On some installations the connection pool doesn't work without this option.
    $mem->setOption(Memcached::OPT_COMPRESSION, false); // if you don't want surprises with a transparent compression.
}
$mem->set('subject', json_encode([
    'event_time' => time(),
    'event_type' => 'type',
    'payload'    => 'XXXX'
]));
```

## [Grafana Dashboard](/grafana/dashboard.json)
![Grafana dashboard](/grafana/dashboard.png)
