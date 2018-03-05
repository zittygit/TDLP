import sys
sys.path.append("/tensorflow/config")
import config

job_name = sys.argv[1]
task_index = int(sys.argv[2])
cluster, server = config.init_cluster(job_name, task_index)
