package main

import (
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
)

// Function to collect and update CPU metrics
func collectCPUMetrics(cpuUsageGauge prometheus.Gauge, cpuIdleGauge prometheus.Gauge, cpuModeGaugeVec *prometheus.GaugeVec) {
	cpuTimes, _ := cpu.Times(true)
	totalTime := cpuTimes[0].User + cpuTimes[0].System + cpuTimes[0].Nice + cpuTimes[0].Idle + cpuTimes[0].Iowait + cpuTimes[0].Irq + cpuTimes[0].Softirq + cpuTimes[0].Steal + cpuTimes[0].Guest + cpuTimes[0].GuestNice

	for _, cpuTime := range cpuTimes {
		cpuModeGaugeVec.WithLabelValues(cpuTime.CPU, "idle").Set(cpuTime.Idle)
		cpuModeGaugeVec.WithLabelValues(cpuTime.CPU, "iowait").Set(cpuTime.Iowait)
		cpuModeGaugeVec.WithLabelValues(cpuTime.CPU, "irq").Set(cpuTime.Irq)
		cpuModeGaugeVec.WithLabelValues(cpuTime.CPU, "nice").Set(cpuTime.Nice)
		cpuModeGaugeVec.WithLabelValues(cpuTime.CPU, "softirq").Set(cpuTime.Softirq)
		cpuModeGaugeVec.WithLabelValues(cpuTime.CPU, "steal").Set(cpuTime.Steal)
		cpuModeGaugeVec.WithLabelValues(cpuTime.CPU, "system").Set(cpuTime.System)
		cpuModeGaugeVec.WithLabelValues(cpuTime.CPU, "user").Set(cpuTime.User)
	}

	// Calculate CPU usage
	cpuUsage := 100.0 * ((totalTime - cpuTimes[0].Idle) / totalTime)

	// Set CPU usage gauge
	cpuUsageGauge.Set(cpuUsage)

	// Set CPU idle gauge
	cpuIdleGauge.Set(100.0 - cpuUsage)
}
func main() {
	// Register Prometheus metrics for memory, CPU, disk count, disk partitions, and network I/O
	totalMemoryGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "total_memory",
		Help: "Total memory available",
	})
	prometheus.MustRegister(totalMemoryGauge)

	freeMemoryGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "free_memory",
		Help: "Free memory available",
	})
	prometheus.MustRegister(freeMemoryGauge)

	usedMemoryGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "used_memory",
		Help: "Used memory",
	})
	prometheus.MustRegister(usedMemoryGauge)

	cpuUsageGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "node_cpu_seconds_total",
		Help: "CPU usage percentage",
	})
	prometheus.MustRegister(cpuUsageGauge)

	cpuModeGaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "node_cpu_seconds",
		Help: "CPU seconds by mode",
	}, []string{"cpu", "mode"})
	prometheus.MustRegister(cpuModeGaugeVec)

	cpuIdleGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cpu_idle",
		Help: "CPU idle percentage",
	})
	prometheus.MustRegister(cpuIdleGauge)

	diskSizeGaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "node_filesystem_size_bytes",
		Help: "Filesystem size in bytes",
	}, []string{"device", "fstype", "mountpoint"})
	prometheus.MustRegister(diskSizeGaugeVec)

	// Register additional memory metrics
	memTotalBytes := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "node_memory_MemTotal_bytes",
		Help: "Total amount of memory in bytes.",
		ConstLabels: prometheus.Labels{
			"unit": "bytes",
		},
	})
	prometheus.MustRegister(memTotalBytes)

	memAvailableBytes := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "node_memory_MemAvailable_bytes",
		Help: "Amount of available memory in bytes.",
		ConstLabels: prometheus.Labels{
			"unit": "bytes",
		},
	})
	prometheus.MustRegister(memAvailableBytes)

	diskAvailGaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "node_filesystem_avail_bytes",
		Help: "Filesystem available space in bytes",
	}, []string{"device", "fstype", "mountpoint"})
	prometheus.MustRegister(diskAvailGaugeVec)

	// Register additional memory metrics
	memInfo, _ := mem.VirtualMemory()
	totalMemoryGauge.Set(float64(memInfo.Total))
	freeMemoryGauge.Set(float64(memInfo.Available))
	usedMemoryGauge.Set(float64(memInfo.Used))

	bufferMemoryGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "node_memory_Buffers_bytes",
		Help: "Buffer memory usage in bytes",
		ConstLabels: prometheus.Labels{
			"unit": "bytes",
		},
	})
	prometheus.MustRegister(bufferMemoryGauge)
	bufferMemoryGauge.Set(float64(memInfo.Buffers))

	cachedMemoryGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "node_memory_Cached_bytes",
		Help: "Cached memory usage in bytes",
		ConstLabels: prometheus.Labels{
			"unit": "bytes",
		},
	})
	prometheus.MustRegister(cachedMemoryGauge)
	cachedMemoryGauge.Set(float64(memInfo.Cached))

	swapMemory, _ := mem.SwapMemory()
	swapTotalMemoryGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "node_memory_SwapTotal_bytes",
		Help: "Total swap memory available in bytes",
		ConstLabels: prometheus.Labels{
			"unit": "bytes",
		},
	})
	prometheus.MustRegister(swapTotalMemoryGauge)
	swapTotalMemoryGauge.Set(float64(swapMemory.Total))

	swapFreeMemoryGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "node_memory_SwapFree_bytes",
		Help: "Free swap memory available in bytes",
		ConstLabels: prometheus.Labels{
			"unit": "bytes",
		},
	})
	prometheus.MustRegister(swapFreeMemoryGauge)
	swapFreeMemoryGauge.Set(float64(swapMemory.Free))

	// Register additional disk metrics
	partitions, _ := disk.Partitions(false)
	for _, partition := range partitions {
		usageStat, _ := disk.Usage(partition.Mountpoint)
		diskSizeGaugeVec.WithLabelValues(partition.Device, partition.Fstype, partition.Mountpoint).Set(float64(usageStat.Total))
		diskAvailGaugeVec.WithLabelValues(partition.Device, partition.Fstype, partition.Mountpoint).Set(float64(usageStat.Free))
	}

	diskReadBytesTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "node_disk_read_bytes_total",
		Help: "Total number of bytes read from disk",
	}, []string{"device"})
	prometheus.MustRegister(diskReadBytesTotal)

	diskWrittenBytesTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "node_disk_written_bytes_total",
		Help: "Total number of bytes written to disk",
	}, []string{"device"})
	prometheus.MustRegister(diskWrittenBytesTotal)

	// Collect disk metrics periodically
	go func() {
		for {
			diskStats, _ := disk.IOCounters()
			for _, stat := range diskStats {
				diskReadBytesTotal.WithLabelValues(stat.Name).Add(float64(stat.ReadBytes))
				diskWrittenBytesTotal.WithLabelValues(stat.Name).Add(float64(stat.WriteBytes))
			}
			time.Sleep(1 * time.Second)
		}
	}()

	// Register additional network metrics
	netStats, _ := net.IOCounters(false)
	for _, stats := range netStats {
		networkReceiveBytesTotal := prometheus.NewCounter(prometheus.CounterOpts{
			Name: "node_network_receive_bytes_total",
			Help: "Total number of bytes received on network interface",
			ConstLabels: prometheus.Labels{
				"interface": stats.Name,
				"unit":      "bytes",
			},
		})
		prometheus.MustRegister(networkReceiveBytesTotal)

		networkTransmitBytesTotal := prometheus.NewCounter(prometheus.CounterOpts{
			Name: "node_network_transmit_bytes_total",
			Help: "Total number of bytes transmitted on network interface",
			ConstLabels: prometheus.Labels{
				"interface": stats.Name,
				"unit":      "bytes",
			},
		})
		prometheus.MustRegister(networkTransmitBytesTotal)

		go func(iface string) {
			for {
				netStat, _ := net.IOCounters(false)
				for _, netInterface := range netStat {
					if netInterface.Name == iface {
						networkReceiveBytesTotal.Add(float64(netInterface.BytesRecv))
						networkTransmitBytesTotal.Add(float64(netInterface.BytesSent))
					}
				}
				time.Sleep(1 * time.Second)
			}
		}(stats.Name)

	}

	// Register process list metric
	processListGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "process_list",
		Help: "List of processes running on the server",
	}, []string{"pid", "name"})
	prometheus.MustRegister(processListGauge)

	// Update process list metric periodically
	go func() {
		for {
			processes, err := getProcessList()
			if err != nil {
				fmt.Println("Error fetching process list:", err)
			} else {
				processListGauge.Reset()
				for _, process := range processes {
					processListGauge.WithLabelValues(process.Pid, process.Name).Set(1)
				}
			}
			time.Sleep(10 * time.Second) // Update process list every 10 seconds
		}
	}()

	go func() {
		for {
			collectCPUMetrics(cpuUsageGauge, cpuIdleGauge, cpuModeGaugeVec)
			time.Sleep(1 * time.Second) // Update CPU metrics every second
		}
	}()

	// Expose the registered metrics via HTTP.
	http.Handle("/metrics", promhttp.Handler())
	fmt.Println("Server is running on port 8083")
	http.ListenAndServe(":8083", nil)
}

// Function to get the list of processes using the `ps` command
func getProcessList() ([]Process, error) {
	cmd := exec.Command("ps", "-e", "-o", "pid,comm")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	processes := make([]Process, 0, len(lines)-1) // excluding header line
	for _, line := range lines[1:] {              // skipping header line
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			processes = append(processes, Process{Pid: fields[0], Name: fields[1]})
		}
	}
	return processes, nil
}

// Process struct to hold process information
type Process struct {
	Pid  string
	Name string
}
