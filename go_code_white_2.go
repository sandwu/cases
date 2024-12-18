func GenerateChart_RedisSingle_BaseResource() {

	// TODO add by zlr, 需要修改， 需要国际化
	GenerateChartSql("redis", "single", "6.0", "data node",
		false,
		"cpu",
		"CPU使用率", "节点的CPU使用率", "%", []string{
			`avg(rate(container_cpu_usage_seconds_total{job="cadvisor", namespace="$namespace", container="redis"}[1m])*100)`,
		},
		[]string{
			"CPU使用率",
		})

	GenerateChartSql("redis", "single", "6.0", "data node",
		false,
		"memory",
		"内存使用量", "节点的内存使用量", "MB", []string{
			`sum(container_memory_usage_bytes{job="cadvisor", namespace="$namespace", container="redis"}/(1024*1024))`,
		},
		[]string{
			"内存使用量",
		})
	GenerateChartSql("redis", "single", "6.0", "data node",
		false,
		"memory",
		"内存使用率", "redis中间件的内存使用率待修改", "%", []string{
			`100*avg(container_memory_usage_bytes{job="cadvisor", namespace="$namespace", container="redis"}/container_spec_memory_limit_bytes{job="cadvisor", namespace="$namespace", container="redis"})`,
		},
		[]string{
			"内存使用率",
		})
	GenerateChartSql("redis", "single", "6.0", "data node",
		false,
		"disk",
		"数据盘写入IOPS", "节点的数据盘写入IOPS", "times/s", []string{
			`sum(rate(container_fs_writes_total{job="cadvisor", cluster="$namespace", device=~"/dev/mapper/.*"}[1m]))`,
		},
		[]string{
			"数据盘写入IOPS",
		})

	GenerateChartSql("redis", "single", "6.0", "data node",
		false,
		"disk",
		"数据盘读取IOPS", "节点的数据盘读取IOPS", "times/s", []string{
			`sum(rate(container_fs_reads_total{job="cadvisor", cluster="$namespace", device=~"/dev/mapper/.*"}[1m]))`,
		},
		[]string{
			"数据盘读取IOPS",
		})

	GenerateChartSql("redis", "single", "6.0", "data node",
		false,
		"disk",
		"数据盘写入吞吐量", "节点的数据盘写入吞吐量", "MB/s", []string{
			`sum(rate(container_fs_writes_bytes_total{job="cadvisor", cluster="$namespace", device=~"/dev/mapper/.*"}/(1024*1024)[1m]))`,
		},
		[]string{
			"数据盘写入吞吐量",
		})

	GenerateChartSql("redis", "single", "6.0", "data node",
		false,
		"disk",
		"数据盘读取吞吐量", "节点的数据盘读取吞吐量", "MB/s", []string{
			`sum(rate(container_fs_reads_bytes_total{job="cadvisor", cluster="$namespace", device=~"/dev/mapper/.*"}/(1024*1024)[1m]))`,
		},
		[]string{
			"数据盘读取吞吐量",
		})

	GenerateChartSql("redis", "single", "6.0", "data node",
		false,
		"disk",
		"数据盘使用量", "节点的数据盘使用量", "GB", []string{
			`sum(container_fs_usage_bytes{job="cadvisor", cluster="$namespace", device=~"/dev/mapper/.*"}/(1024*1024*1024))`,
		},
		[]string{
			"数据盘使用量",
		})

	GenerateChartSql("redis", "single", "6.0", "data node",
		false,
		"disk",
		"数据盘使用率", "节点的数据盘使用率", "%", []string{
			`avg(100*container_fs_usage_bytes{job="cadvisor", cluster="$namespace", device=~"/dev/mapper/.*"}/container_fs_limit_bytes{job="cadvisor", cluster="$namespace", device=~"/dev/mapper/.*"})`,
		},
		[]string{
			"数据盘使用率",
		})
}