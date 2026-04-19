CREATE TABLE IF NOT EXISTS `calls_queue` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `lead` varchar(255) NOT NULL,
  `branch` varchar(255) NOT NULL,
  `dst` varchar(255) NOT NULL,
  `src` varchar(255) NOT NULL,
  `ivr` varchar(255) NOT NULL,
  `type` int(1) NOT NULL DEFAULT 1,
  `retry_time` datetime NOT NULL DEFAULT current_timestamp(),
  `retries_left` int(11) NOT NULL DEFAULT 3,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `dialer_branches` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `branch` varchar(255) NOT NULL DEFAULT '',
  `start_time` datetime DEFAULT current_timestamp(),
  `stop_time` datetime DEFAULT current_timestamp(),
  `rows_processed` int(11) DEFAULT NULL,
  `rows_total` int(11) DEFAULT NULL,
  `rows_ok` int(11) DEFAULT NULL,
  `parent_id` int(11) DEFAULT NULL,
  `comment` text DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `geos` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `geo` varchar(80) NOT NULL,
  `geo2` varchar(2) NOT NULL,
  `code` varchar(20) NOT NULL,
  `prefix` varchar(20) DEFAULT NULL,
  `src` varchar(255) NOT NULL,
  `provider` varchar(255) NOT NULL,
  `nlines` int(11) NOT NULL DEFAULT 30,
  `cps` int(11) NOT NULL DEFAULT 10,
  `comment` text DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `queue_log` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `time` varchar(32) DEFAULT NULL,
  `callid` char(64) DEFAULT NULL,
  `queuename` char(64) DEFAULT NULL,
  `agent` char(64) DEFAULT NULL,
  `event` char(32) DEFAULT NULL,
  `data` char(64) DEFAULT NULL,
  `data1` char(64) DEFAULT NULL,
  `data2` char(64) DEFAULT NULL,
  `data3` char(64) DEFAULT NULL,
  `data4` char(64) DEFAULT NULL,
  `data5` char(64) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `queues` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `Queue` varchar(255) NOT NULL,
  `Available` int(11) NOT NULL DEFAULT 0,
  `Callers` int(11) NOT NULL DEFAULT 0,
  `HoldTime` int(11) NOT NULL DEFAULT 0,
  `LoggedIn` int(11) NOT NULL DEFAULT 0,
  `LongestHoldTime` int(11) NOT NULL DEFAULT 0,
  `TalkTime` int(11) NOT NULL DEFAULT 0,
  `astrerisk` varchar(255) NOT NULL DEFAULT 'none',
  PRIMARY KEY (`id`),
  KEY `Queuename` (`Queue`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `tasks` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `phone_number` longtext NOT NULL,
  `nlines` int(8) DEFAULT 30,
  `cps` int(8) DEFAULT 10,
  `ivr` varchar(80) DEFAULT NULL,
  `retries` int(1) NOT NULL DEFAULT 3,
  `call_timeout` int(4) NOT NULL DEFAULT 60000,
  `retry_time` int(4) NOT NULL DEFAULT 1800,
  `uid` varchar(80) DEFAULT NULL,
  `type` int(1) NOT NULL DEFAULT 1 COMMENT '1 for autoinformer/2 progressive',
  `ready` int(1) NOT NULL DEFAULT 0,
  `gmt` varchar(4) NOT NULL DEFAULT '+0',
  `stoptime` varchar(12) NOT NULL DEFAULT '19',
  `starttime` varchar(12) NOT NULL DEFAULT '9',
  `stop` datetime DEFAULT NULL,
  `start` datetime DEFAULT NULL,
  `geo_id` int(11) NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `tasks_ivr` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  `queue_name` varchar(255) NOT NULL DEFAULT '',
  `type` int(1) NOT NULL DEFAULT 1,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `users` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `username` varchar(255) NOT NULL,
  `password` varchar(255) NOT NULL,
  `accesslevel` int(11) NOT NULL DEFAULT 100,
  `access` varchar(255) NOT NULL DEFAULT 'UAELD',
  `comment` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `cdr` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `calldate` datetime DEFAULT NULL,
  `answerdate` datetime DEFAULT NULL,
  `enddate` datetime DEFAULT NULL,
  `clid` varchar(80) NOT NULL DEFAULT '',
  `src` varchar(80) NOT NULL DEFAULT '',
  `dst` varchar(80) NOT NULL DEFAULT '',
  `currentopnumber` varchar(80) NOT NULL DEFAULT '',
  `prefix` varchar(80) NOT NULL DEFAULT '',
  `dcontext` varchar(80) NOT NULL DEFAULT '',
  `channel` varchar(80) NOT NULL DEFAULT '',
  `in_trunk` varchar(80) NOT NULL DEFAULT '',
  `dstchannel` varchar(80) NOT NULL DEFAULT '',
  `out_trunk` varchar(80) NOT NULL DEFAULT '',
  `lastapp` varchar(80) NOT NULL DEFAULT '',
  `lastdata` varchar(80) NOT NULL DEFAULT '',
  `duration` int(11) NOT NULL DEFAULT 0,
  `billsec` int(11) NOT NULL DEFAULT 0,
  `holdsec` int(11) NOT NULL DEFAULT 0,
  `disposition` varchar(45) NOT NULL DEFAULT '',
  `amaflags` int(11) NOT NULL DEFAULT 0,
  `accountcode` varchar(20) NOT NULL DEFAULT '',
  `uniqueid` varchar(32) NOT NULL DEFAULT '',
  `linkedid` varchar(32) NOT NULL DEFAULT '',
  `userfield` varchar(255) NOT NULL DEFAULT '',
  `recordingstorage` varchar(100) NOT NULL DEFAULT '',
  `recordingfile` varchar(255) NOT NULL DEFAULT '',
  `type` varchar(80) NOT NULL DEFAULT '',
  `department` varchar(80) NOT NULL DEFAULT '',
  `country` varchar(80) NOT NULL DEFAULT '',
  `provider` varchar(80) NOT NULL DEFAULT '',
  `ivrstatus` varchar(40) NOT NULL DEFAULT '',
  `hunguped` varchar(80) NOT NULL DEFAULT '',
  `queueofpeople` int(11) NOT NULL DEFAULT 0,
  `sip_callid` varchar(100) NOT NULL DEFAULT '',
  `sip_cause` varchar(50) NOT NULL DEFAULT '',
  `sip_cause_desc` varchar(80) NOT NULL DEFAULT '',
  `hangupcause` int(11) NOT NULL DEFAULT 0,
  `hangupsource` varchar(80) NOT NULL DEFAULT '',
  `info` varchar(255) DEFAULT NULL,
  `ivr` varchar(80) DEFAULT NULL,
  `stt` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `calldate` (`calldate`),
  KEY `dst` (`dst`),
  KEY `accountcode` (`accountcode`),
  KEY `uniqueid` (`uniqueid`),
  KEY `sip_callid` (`sip_callid`),
  KEY `userfield` (`userfield`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO `users` (`id`, `username`, `password`, `accesslevel`, `access`, `comment`)
VALUES
  (1, 'admin', '$2a$10$rj1bFVRiOVwpyNMBdJxkROnEdHBcW7iMVkpYHjMl6ijAtBT9K.szC', 100, 'UAL', 'local compose admin')
ON DUPLICATE KEY UPDATE
  `username` = VALUES(`username`),
  `password` = VALUES(`password`),
  `accesslevel` = VALUES(`accesslevel`),
  `access` = VALUES(`access`),
  `comment` = VALUES(`comment`);
