/*
Navicat MySQL Data Transfer

Source Server         : 本地
Source Server Version : 50621
Source Host           : localhost:3306
Source Database       : deepnex_dev

Target Server Type    : MYSQL
Target Server Version : 50621
File Encoding         : 65001

Date: 2017-12-28 21:41:18
*/

SET FOREIGN_KEY_CHECKS=0;

-- ----------------------------
-- Table structure for `alembic_version`
-- ----------------------------
DROP TABLE IF EXISTS `alembic_version`;
CREATE TABLE `alembic_version` (
  `version_num` varchar(32) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- ----------------------------
-- Records of alembic_version
-- ----------------------------
INSERT INTO `alembic_version` VALUES ('e048fbacef76');

-- ----------------------------
-- Table structure for `applications`
-- ----------------------------
DROP TABLE IF EXISTS `applications`;
CREATE TABLE `applications` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `aid` int(11) DEFAULT NULL,
  `name` varchar(128) NOT NULL,
  `info` text NOT NULL,
  `param` text NOT NULL,
  `cover_img` varchar(128) DEFAULT NULL,
  `param_guide` text,
  `path` varchar(128) DEFAULT NULL,
  `updatedTime` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=14 DEFAULT CHARSET=utf8;

-- ----------------------------
-- Records of applications
-- ----------------------------
INSERT INTO `applications` VALUES ('6', '1', 'Spark集群', 'Spark是专为大规模数据处理而设计的快速通用的计算引擎', '[{\"note\": \"spark\\u8282\\u70b9\\u6570\", \"type\": \"int\", \"name\": \"nodes\"}]', '/static/upload/application/6.png?t=1513580298.97', 'Spark 模板的参数包括CPU个数、内存大小以及节点数', 'bin/spark', '2017-12-08 11:06:55');
INSERT INTO `applications` VALUES ('7', '2', 'zeppelin', 'spark的前端展示框架', '[{\"note\": \"spark\\u7684\\u8def\\u5f84\", \"type\": \"varchar\", \"name\": \"spark\"}]', '/static/upload/application/7.png?t=1513581878.41', '需要填写spark的路径', 'bin/zeppelin', '2017-12-08 11:12:21');
INSERT INTO `applications` VALUES ('8', '3', 'Tensorflow_CPU版', '一个采用数据流图用于数值计算的开源软件库', '[]', '/static/upload/application/8.png?t=1513582227.62', '只需要填写CPU与内存大小', 'bin/tensorflow-cpu', '2017-12-08 11:14:47');
INSERT INTO `applications` VALUES ('9', '4', 'Tensorflow_CPU集群版', '一个采用数据流图用于数值计算的开源软件库', '[{\"note\": \"\\u53c2\\u6570\\u670d\\u52a1\\u5668\\u4e2a\\u6570\", \"type\": \"int\", \"name\": \"parameterservers\"}, {\"note\": \"\\u5de5\\u4f5c\\u670d\\u52a1\\u5668\\u4e2a\\u6570\", \"type\": \"int\", \"name\": \"workerservers\"}]', '/static/upload/application/9.png?t=1513582246.06', '需要填写参数服务器和工作服务器的数量', 'bin/tensorflow-cpu-cluster', '2017-12-08 11:20:05');
INSERT INTO `applications` VALUES ('10', '7', 'Hadoop集群', 'Hadoop是一个由Apache基金会所开发的分布式系统基础架构', '[{\"note\": \"\\u8282\\u70b9\\u6570\", \"type\": \"int\", \"name\": \"nodes\"}]', '/static/upload/application/10.png?t=1513582334.53', '需要填写节点数', 'bin/hadoop', '2017-12-08 11:21:47');
INSERT INTO `applications` VALUES ('13', '21', 'TensorFlow-GPU', 'tensorflow-gpu', '[{\"note\": \"\\u5fc5\\u987b\\u662f\\u6574\\u6570\\u503c\", \"type\": \"int\", \"name\": \"gpu\"}]', '/static/upload/application/13.png?t=1513582254.0', 'gpu的数量必须是整数值', 'bin/tensorflow-gpu', '2017-12-12 14:26:50');

-- ----------------------------
-- Table structure for `news`
-- ----------------------------
DROP TABLE IF EXISTS `news`;
CREATE TABLE `news` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `title` varchar(128) NOT NULL,
  `content` text NOT NULL,
  `visitNum` int(11) DEFAULT NULL,
  `updatedTime` datetime DEFAULT NULL,
  `poster` varchar(128) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8;

-- ----------------------------
-- Records of news
-- ----------------------------
INSERT INTO `news` VALUES ('6', '2017“天河二号”超算用户年会完美落幕！', '<p>\r\r2017年12月16日，为期两天的国家超算广州中心2017“天河二号”用户年会载满各领域应用创新成果和对未来发展的美好期待完美落幕。会议期间，来自全国各地高校、科研院所和企业的320多位用户代表齐聚一堂，交流观点，切磋技术，分享超算系统技术的最新进展和领域应用的创新成果，充分展现我国超算应用的蓬勃发展态势。会议期间，国家超算广州中心正式发布全新的天河星光云超算平台2.0，是我中心始终坚持技术应用服务三位一体发展、不断提升创新能力的又一重要举措。\r\r<br></p><p><br></p><p><img alt=\"\" src=\"http://www.nscc-gz.cn/userfiles/files/%E7%AC%AC2%E7%AF%87%EF%BC%9A2017%E2%80%9C%E5%A4%A9%E6%B2%B3%E4%BA%8C%E5%8F%B7%E2%80%9D%E8%B6%85%E7%AE%97%E7%94%A8%E6%88%B7%E5%B9%B4%E4%BC%9A%E5%AE%8C%E7%BE%8E%E8%90%BD%E5%B9%95%EF%BC%81/1.jpg\"><br></p>', '1', '2017-12-27 15:15:56', '广州超算中心');
INSERT INTO `news` VALUES ('7', '\"天河二号”云超算与大数据处理技术高级研修班', '<p>\r\r</p><p>根据《人力资源社会保障部办公厅关于印发专业技术人才知识更新工程2017年高级研修项目计划的通知》（人社厅发〔2017〕36号），“天河二号”云超算与大数据处理技术高级研修项目已经国家人力资源社会保障部办公厅审核批准。为发挥“天河二号”重大科技基础设施的功能，提升科研技术人员使用超算的能力，以超级计算机支撑产业升级，推进国家科技创新发展，提高综合国力和国家科技竞争力，在广东省人力资源和社会保障厅的大力支持下，中山大学国家超级计算广州中心将于2017年6月29至7月4日举办《“天河二号”云超算技术与大数据处理技术高级研修班》活动。热忱欢迎各位学员参加本次高级研修班！</p><p>&nbsp; &nbsp; &nbsp; &nbsp;现将有关事项通知如下。</p><p>一、 &nbsp; &nbsp;<strong>研修对象</strong></p><p>本次研修班面向全国各省（自治区、直辖市）从事高性能计算、云计算、大数据、机器学习相关工作实践的<strong>高层次专业技术人才或高级管理人才</strong>，共招收70名学员（各省市的招收名额分配表见<strong>附件1</strong>）。</p><p>二、 &nbsp; &nbsp;<strong>研修内容及研修方式</strong></p><p>（一）<strong>研修内容</strong></p><p>1、 &nbsp; &nbsp;中国超级计算发展的机遇与挑战；</p><p>2、 &nbsp; &nbsp;高性能计算与大数据融合创新发展；</p><p>3、 &nbsp; &nbsp;“天河二号”的云超算平台；</p><p>4、 &nbsp; &nbsp;高性能计算技术；</p><p>5、 &nbsp; &nbsp;云计算理论与技术；</p><p>6、 &nbsp; &nbsp;超算中心云平台介绍及实践；</p><p>7、 &nbsp; &nbsp;大数据技术及应用；</p><p>8、 &nbsp; &nbsp;基于超算的大数据处理及实践；</p><p>9、 &nbsp; &nbsp;深度学习技术及应用；</p><p>10、 &nbsp; &nbsp;“天河二号”主机系统机房参观及讨论。</p><p>（二）<strong>研修方式</strong></p><p>主要采取专家授课、学员讨论、上机实践、案例教学与实地考察相结合的研修方式，做到理论联系实际，讲求实效。</p><p>三、<strong>组织机构</strong></p><p><strong>主办单位</strong>： 广东省人力资源和社会保障厅</p><p><strong>承办单位</strong>：中山大学国家超级计算广州中心</p>\r\r<br><p></p>', '0', '2017-12-27 15:16:36', '广州超算中心');
INSERT INTO `news` VALUES ('8', '学术报告：信息融合技术', '<p>\r\r</p><p><strong>报告题目</strong>：信息融合技术</p><p><strong>主讲</strong>： 史习智 &nbsp;教授</p><p><strong>日期</strong>：2017 年12 月12日 (周二)</p><p><strong>时间</strong>：下午14：30-16：30</p><p><strong>地点</strong>：中山大学东校区数据科学与计算机学院A101</p><p><strong>主持</strong>：卓汉逵 &nbsp;副教授</p><p>&nbsp;</p><p>&nbsp;</p><p><strong>摘要</strong>:</p><p>信息融合技术利用计算机技术对按时序获得的多维时空观测信息在一定准则下加以自动分析、综合处理，以完成所需的决策和估计任务而进行的信息处理过程。报告将分别介绍三方面的内容，即跟踪和多传感器数据融合、分布式融合和情景增强融合。</p><p>&nbsp;</p><p>&nbsp;</p><p><strong>报告人简介</strong>:</p><p>史习智，上海交通大学教授（退休）。</p>\r\r<br><p></p>', '0', '2017-12-27 15:17:31', '中山大学');
INSERT INTO `news` VALUES ('9', '学术报告：宝洁如何用大数据引领商业变革', '<p>\r\r</p><p><strong>报告题目</strong>：宝洁如何用大数据引领商业变革</p><p><strong>主讲</strong>：Guy Peri （P&amp;G首席数据官）</p><p><strong>日期</strong>： 2017年12月5日</p><p><strong>时间</strong>： 15:00-17:00</p><p><strong>地点</strong>：广州大学城中山大学数据科学与计算机学院A101讲学厅</p><p>&nbsp;</p><p><strong>摘要</strong>:</p><p>数据架构设计 ？机器学习？深度学习？算法模型，人工智能这些每天出现在课堂上、课本上的知识点，到底是如何在企业中发挥作用、创造商机和价值的？</p><p>这一次！ 我们请到了宝洁全球副总裁—首席数据官 Guy Peri！为我们揭开世界最大的快消品公司，如何利用大数据技术洞察生意的面纱！同时，也给你带来一个完全颠覆你认知的“IT-信息技术部”！</p><p><strong>讲座主要内容：</strong></p><p>宝洁全球大数据的商业应用现状</p><p>宝洁大数据战略及在产品创新、业务领域和供应链的应用</p><p>宝洁全球大数据分析成功案例分享</p><p>宝洁信息技术人才如何引领未来的商业变革</p><p>宝洁数据科学家团队职场分享</p><p><strong>报告人简介</strong>:</p><p><strong>Guy Peri</strong></p><p><strong>&nbsp; </strong><img alt=\"\"></p><p><strong>宝洁全球信息技术部副总裁、首席数据官</strong>&nbsp;<strong>Vice President</strong><strong>，</strong><strong>Chief Data Officer</strong><strong>Information Technology at Procter &amp; Gamble</strong></p><p>全球大数据分析行业首批数据科学家（CDO），荣登哈佛商业周刊（How P&amp;G and American Express Are Approaching AI）</p><p>被“The International Institute of Analytics”授予“Leadership in Analytics”荣誉称号</p><p>被授予美国Top 100 CIO荣誉，荣登“计算机世界”头条</p><p>为哈佛商学院的学生讲授宝洁大数据分析商业案例</p>\r\r<br><p></p>', '3', '2017-12-27 15:18:09', '中山大学');

-- ----------------------------
-- Table structure for `notices`
-- ----------------------------
DROP TABLE IF EXISTS `notices`;
CREATE TABLE `notices` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `title` varchar(128) NOT NULL,
  `content` text NOT NULL,
  `visitNum` int(11) DEFAULT NULL,
  `updatedTime` datetime DEFAULT NULL,
  `poster` varchar(128) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8;

-- ----------------------------
-- Records of notices
-- ----------------------------
INSERT INTO `notices` VALUES ('2', '欢迎大家使用天河云深度学习平台', '<p>欢迎大家使用天河云深度学习平台！<br></p><p><br></p><p>请先完善自己的个人资料，等待管理员进行审核。</p><p><br></p><p>审核通过后即可使用集群应用模板。</p>', '0', '2017-12-27 15:09:24', '系统官方');
INSERT INTO `notices` VALUES ('3', '关于用户等级权限的说明', '<p>略</p>', '0', '2017-12-27 15:10:12', '系统官方');
INSERT INTO `notices` VALUES ('4', '关于用户向系统官方反馈的途径', '<p>略</p>', '0', '2017-12-27 15:10:42', '系统官方');
INSERT INTO `notices` VALUES ('5', '近期将推出一系列的深度学习课程', '<p>如题</p>', '0', '2017-12-27 15:11:10', '系统官方');

-- ----------------------------
-- Table structure for `users`
-- ----------------------------
DROP TABLE IF EXISTS `users`;
CREATE TABLE `users` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `username` varchar(128) NOT NULL,
  `password_hash` varchar(128) NOT NULL,
  `email` varchar(128) NOT NULL,
  `description` varchar(128) DEFAULT NULL,
  `real_name` varchar(128) DEFAULT NULL,
  `phone` varchar(128) DEFAULT NULL,
  `address` varchar(128) DEFAULT NULL,
  `last_login` datetime DEFAULT NULL,
  `date_joined` datetime DEFAULT NULL,
  `permissions` int(11) NOT NULL,
  `is_auth` int(11) NOT NULL,
  `avatar_url` varchar(128) DEFAULT NULL,
  `api_password` varchar(128) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `ix_users_email` (`email`),
  UNIQUE KEY `ix_users_username` (`username`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8;

-- ----------------------------
-- Records of users
-- ----------------------------
INSERT INTO `users` VALUES ('1', 'long', 'pbkdf2:sha1:1000$m5sEKr8n$4c5726f7403c6a70b6f06cb9fd7ce66561aba747', 'admin@qq.com', null, '邹哲鹏', '13802401911', '中山大学', '2017-12-26 20:39:09', '2017-11-09 10:07:47', '0', '1', '/static/upload/avatar/1.png?t=1514030154.8', '123456');
INSERT INTO `users` VALUES ('4', 'zhuanglei', 'pbkdf2:sha256:50000$iTobiAEb$4d293f301de54409dcbd8908de62f05a9205f3b749a32b9b898bac150c1a841f', '2697950380@qq.com', null, '庄磊', '13802411912', '中山大学', '2017-12-15 16:02:34', '2017-12-15 15:45:32', '1', '0', '/static/resource/img/none.jpg', 'Uo1e5Q7j');
