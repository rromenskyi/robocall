/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;


CREATE TABLE `calls_queue` (
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `dialer_branches` (
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `geos` (
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `queue_log` (
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `queues` (
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `tasks` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `phone_number` longtext NOT NULL DEFAULT 'NULL',
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `tasks_ivr` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  `queue_name` varchar(255) NOT NULL DEFAULT '',
  `type` int(1) NOT NULL DEFAULT 1,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `users` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `username` varchar(255) NOT NULL,
  `password` varchar(255) NOT NULL,
  `accesslevel` int(11) NOT NULL DEFAULT 100,
  `access` varchar(255) NOT NULL DEFAULT 'UAELD',
  `comment` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

INSERT INTO `users` (`id`, `username`, `password`, `accesslevel`, `access`, `comment`) VALUES
(1, 'admin', '$2a$10$BZ1ZONBO4/gQPCEx84kx6uJtA0/Edm5lUf.cVMTE..wejgtKPwiZu', 100, 'UAL', 'example@example.com');


/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;