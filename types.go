package main

import "time"

type IpRtt struct {
	Ip  string
	Rtt time.Duration
}

type IpRtts []IpRtt

func (i IpRtts) Len() int {
	return len(i)
}

func (i IpRtts) Less(j, k int) bool {
	return i[j].Rtt < i[k].Rtt
}

func (i IpRtts) Swap(j, k int) {
	i[j], i[k] = i[k], i[j]
}

type Meta struct {
	Message string   `json:"message"`
	Web     []string `json:"web"`
}
