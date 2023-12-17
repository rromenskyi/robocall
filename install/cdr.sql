/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;


CREATE TABLE `cdr` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `calldate` datetime NOT NULL DEFAULT '0000-00-00 00:00:00',
  `answerdate` datetime NOT NULL DEFAULT '0000-00-00 00:00:00',
  `enddate` datetime NOT NULL DEFAULT '0000-00-00 00:00:00',
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



/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;