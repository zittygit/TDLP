import sys
sys.path.append("/tensorflow/config")
import config
import math
import tensorflow as tf
from tensorflow.examples.tutorials.mnist import input_data

job_name = sys.argv[1]
task_index = int(sys.argv[2])
cluster, server = config.init_cluster(job_name, task_index)
IMAGE_PIXELS = 28
HIDDEN_UNITS = 100
BATCH_SIZE = 200

def main(_):
	with tf.device(tf.train.replica_device_setter(worker_device = "/job:worker/task:%d" % task_index, cluster = cluster)):
		hid_w = tf.Variable(tf.truncated_normal([IMAGE_PIXELS * IMAGE_PIXELS, HIDDEN_UNITS], stddev = 1.0 / IMAGE_PIXELS), name = "hid_w")
		hid_b = tf.Variable(tf.zeros([HIDDEN_UNITS]), name = "hid_b")
		sm_w = tf.Variable(tf.truncated_normal([HIDDEN_UNITS, 10],  stddev = 1.0 / math.sqrt(HIDDEN_UNITS)), name = "sm_w")
		sm_b = tf.Variable(tf.zeros([10]), name = "sm_b")
		x = tf.placeholder(tf.float32, [None, IMAGE_PIXELS * IMAGE_PIXELS])
		y_ = tf.placeholder(tf.float32, [None, 10])
		hid_lin = tf.nn.xw_plus_b(x, hid_w, hid_b)
		hid = tf.nn.relu(hid_lin)
		y = tf.nn.softmax(tf.nn.xw_plus_b(hid, sm_w, sm_b))
		loss = -tf.reduce_sum(y_ * tf.log(tf.clip_by_value(y, 1e-10, 1.0)))
		global_step = tf.Variable(0)
		train_op = tf.train.AdagradOptimizer(0.01).minimize(loss, global_step = global_step)
		correct_prediction = tf.equal(tf.argmax(y, 1), tf.argmax(y_, 1))
		accuracy = tf.reduce_mean(tf.cast(correct_prediction, tf.float32))
		saver = tf.train.Saver()
		summary_op = tf.summary.merge_all()
		init_op = tf.global_variables_initializer()
	sv = tf.train.Supervisor(is_chief = (task_index == 0), logdir = "/tensorflow/logs", init_op = init_op, summary_op = summary_op, saver = saver, global_step = global_step, save_model_secs = 600)
	mnist = input_data.read_data_sets("/tensorflow/data", one_hot = True)
	with sv.managed_session(server.target) as sess:
		step = 0
		while not sv.should_stop() and step < 2000:
			batch_xs, batch_ys = mnist.train.next_batch(BATCH_SIZE)
			train_feed = {x: batch_xs, y_: batch_ys}
			_, step = sess.run([train_op, global_step], feed_dict=train_feed)
			if step % 200 == 0: 
				print "Done step %d" % step
				print(sess.run(accuracy, feed_dict = {x: mnist.test.images, y_: mnist.test.labels}))  
		sess.close()    	
	sv.stop()

if __name__ == "__main__":
	tf.app.run()
