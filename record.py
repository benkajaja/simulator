import GPUtil
import psutil
import time
import logging

LOG_FORMAT = "%(asctime)s %(message)s"
DATE_FORMAT = "%Y/%m/%d %H:%M:%S"

logging.basicConfig(level=logging.DEBUG, format=LOG_FORMAT, datefmt=DATE_FORMAT)

pre_bits_recv = psutil.net_io_counters().bytes_recv * 8
pre_bits_sent = psutil.net_io_counters().bytes_sent * 8

time.sleep(1)

cur_bits_recv = psutil.net_io_counters().bytes_recv * 8
cur_bits_sent = psutil.net_io_counters().bytes_sent * 8

GPUtarget = GPUtil.getGPUs()[0]

log = "[DEBUG] STATS. CPU: %.2f, MEM: %.2f, GPU: %.2f, GPUMEM: %.2f, UPLINK: %.2f, DOWNLINK: %.2f" % (
    psutil.cpu_percent()/100,
    psutil.virtual_memory()[2]/100,
    GPUtarget.load,
    GPUtarget.memoryUtil,
    (cur_bits_sent - pre_bits_sent) /(1024), # in Kbps
    (cur_bits_recv - pre_bits_recv) /(1024) # in Kbps
)

# print(log)
logging.debug(log)