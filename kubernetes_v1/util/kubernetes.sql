-- MySQL dump 10.14  Distrib 5.5.56-MariaDB, for Linux (x86_64)
--
-- Host: localhost    Database: kubernetes
-- ------------------------------------------------------
-- Server version	5.5.56-MariaDB

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `app`
--

DROP TABLE IF EXISTS `app`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `app` (
  `aid` int(8) unsigned NOT NULL AUTO_INCREMENT,
  `appname` varchar(128) NOT NULL,
  `path` varchar(128) NOT NULL,
  `info` text NOT NULL,
  PRIMARY KEY (`aid`),
  UNIQUE KEY `appname` (`appname`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `app`
--

LOCK TABLES `app` WRITE;
/*!40000 ALTER TABLE `app` DISABLE KEYS */;
/*!40000 ALTER TABLE `app` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `config`
--

DROP TABLE IF EXISTS `config`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `config` (
  `cid` int(8) unsigned NOT NULL AUTO_INCREMENT,
  `iid` int(8) unsigned NOT NULL,
  `starttime` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `endtime` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
  `nodes` int(8) unsigned NOT NULL,
  PRIMARY KEY (`cid`),
  KEY `iid` (`iid`),
  CONSTRAINT `config_iid` FOREIGN KEY (`iid`) REFERENCES `instance` (`iid`) ON DELETE CASCADE ON UPDATE NO ACTION
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `config`
--

LOCK TABLES `config` WRITE;
/*!40000 ALTER TABLE `config` DISABLE KEYS */;
/*!40000 ALTER TABLE `config` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `instance`
--

DROP TABLE IF EXISTS `instance`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `instance` (
  `iid` int(8) unsigned NOT NULL AUTO_INCREMENT,
  `instancename` varchar(128) NOT NULL,
  `aid` int(8) unsigned DEFAULT NULL,
  `uid` int(8) unsigned NOT NULL,
  `cid` int(8) unsigned NOT NULL,
  `state` tinyint(1) NOT NULL DEFAULT '0',
  `cpu` int(10) unsigned NOT NULL,
  `memory` int(10) unsigned NOT NULL,
  `createtime` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `deletetime` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
  PRIMARY KEY (`iid`),
  KEY `instancename` (`instancename`),
  KEY `aid` (`aid`),
  KEY `uid` (`uid`),
  CONSTRAINT `instance_aid` FOREIGN KEY (`aid`) REFERENCES `app` (`aid`) ON DELETE SET NULL ON UPDATE NO ACTION,
  CONSTRAINT `instance_uid` FOREIGN KEY (`uid`) REFERENCES `user` (`uid`) ON DELETE CASCADE ON UPDATE NO ACTION
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `instance`
--

LOCK TABLES `instance` WRITE;
/*!40000 ALTER TABLE `instance` DISABLE KEYS */;
/*!40000 ALTER TABLE `instance` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `proxy`
--

DROP TABLE IF EXISTS `proxy`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `proxy` (
  `pid` int(8) unsigned NOT NULL AUTO_INCREMENT,
  `proxyname` varchar(128) NOT NULL,
  `iid` int(8) unsigned NOT NULL,
  `firstport` int(8) unsigned NOT NULL,
  `secondport` int(8) unsigned NOT NULL,
  `httpurl` varchar(256) NOT NULL,
  `websocketurl` varchar(256) NOT NULL,
  PRIMARY KEY (`pid`),
  KEY `iid` (`iid`),
  CONSTRAINT `proxy_iid` FOREIGN KEY (`iid`) REFERENCES `instance` (`iid`) ON DELETE CASCADE ON UPDATE NO ACTION
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `proxy`
--

LOCK TABLES `proxy` WRITE;
/*!40000 ALTER TABLE `proxy` DISABLE KEYS */;
/*!40000 ALTER TABLE `proxy` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `user`
--

DROP TABLE IF EXISTS `user`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `user` (
  `uid` int(8) unsigned NOT NULL AUTO_INCREMENT,
  `username` varchar(128) NOT NULL,
  `gid` int(8) unsigned DEFAULT NULL,
  `role` varchar(8) NOT NULL DEFAULT '',
  `email` varchar(128) NOT NULL,
  `createtime` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`uid`),
  UNIQUE KEY `username` (`username`),
  KEY `gid` (`gid`),
  CONSTRAINT `user_gid` FOREIGN KEY (`gid`) REFERENCES `usergroup` (`gid`) ON DELETE SET NULL ON UPDATE NO ACTION
) ENGINE=InnoDB AUTO_INCREMENT=1001 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `user`
--

LOCK TABLES `user` WRITE;
/*!40000 ALTER TABLE `user` DISABLE KEYS */;
/*!40000 ALTER TABLE `user` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `usergroup`
--

DROP TABLE IF EXISTS `usergroup`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `usergroup` (
  `gid` int(8) unsigned NOT NULL AUTO_INCREMENT,
  `groupname` varchar(128) NOT NULL,
  PRIMARY KEY (`gid`),
  UNIQUE KEY `groupname` (`groupname`)
) ENGINE=InnoDB AUTO_INCREMENT=1001 DEFAULT CHARSET=utf8;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `usergroup`
--

LOCK TABLES `usergroup` WRITE;
/*!40000 ALTER TABLE `usergroup` DISABLE KEYS */;
/*!40000 ALTER TABLE `usergroup` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2017-09-30  9:41:01
