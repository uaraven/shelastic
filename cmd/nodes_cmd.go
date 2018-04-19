package cmd

import (
	"bytes"
	"fmt"
	"shelastic/es"
	"strconv"
	"strings"

	ishell "gopkg.in/abiosoft/ishell.v2"
)

// Nodes wraps all nodes functions
func Nodes() *ishell.Cmd {
	nodes := &ishell.Cmd{
		Name: "node",
		Help: "Node operations",
	}

	nodes.AddCmd(&ishell.Cmd{
		Name: "stats",
		Help: "Get node statistics",
		Func: getNodeStats,
	})

	nodes.AddCmd(&ishell.Cmd{
		Name: "environment",
		Help: "Get OS and JVM statistics",
		Func: getEnvironmentStats,
	})

	nodes.AddCmd(&ishell.Cmd{
		Name: "shards",
		Help: "Show shard allocation for nodes",
		Func: nodeShards,
	})

	nodes.AddCmd(&ishell.Cmd{
		Name: "decomission",
		Help: "Decomission node(s). Usage: decomission [--selector node|ip|host  --clear|list-to-decomission]",
		Func: decomissionNodes,
	})

	return nodes
}

func getNodeStats(c *ishell.Context) {
	if context != nil {
		nodeStats, err := context.GetNodeStats(c.Args)
		if err != nil {
			errorMsg(c, err.Error())
			return
		}
		if err != nil {
			errorMsg(c, err.Error())
			return
		}
		var nodeName string
		if len(c.Args) > 0 {
			nodeName = c.Args[0]
		} else {
			nodeName = ""
		}
		if nodeName != "" {
			for node := range nodeStats.Nodes {
				if node == nodeName {
					cprintln(c, undr(node+":"))
					cprintln(c, nodeStatsToString(nodeStats.Nodes[node]))
					return
				}
			}
			errorMsg(c, "Node '%s' not found", nodeName)
		} else {
			print(c, nodeStats)
		}
	} else {
		errorMsg(c, errNotConnected)
	}
}

func getEnvironmentStats(c *ishell.Context) {
	if context != nil {
		nodeStats, err := context.GetNodeEnvironmentInfo(c.Args)
		if err != nil {
			errorMsg(c, err.Error())
			return
		}
		if err != nil {
			errorMsg(c, err.Error())
			return
		}
		var nodeName string
		if len(c.Args) > 0 {
			nodeName = c.Args[0]
		} else {
			nodeName = ""
		}
		if nodeName != "" {
			for node := range nodeStats.Nodes {
				if node == nodeName {
					cprintln(c, undr(node+":"))
					cprintln(c, environmentToString(nodeStats.Nodes[node]))
					return
				}
			}
			errorMsg(c, "Node '%s' not found", nodeName)
		} else {
			printEnvironment(c, nodeStats.Nodes)
		}
	} else {
		errorMsg(c, errNotConnected)
	}
}

func nodeShards(c *ishell.Context) {
	if context == nil {
		errorMsg(c, errNotConnected)
		return
	}

	var nodes []string

	if len(c.Args) < 1 {
		nodes = make([]string, len(context.Nodes))
		i := 0
		for n := range context.Nodes {
			nodes[i] = context.Nodes[n].Name
			i++
		}
	} else {
		nodes = []string{c.Args[0]}
	}
	indices, err := context.ListIndices()
	if err != nil {
		errorMsg(c, err.Error())
		return
	}

	for _, node := range nodes {
		nodeIndexMap, err := buildNodeIndexInfo(node, indices)
		if err != nil {
			errorMsg(c, err.Error())
			return
		}
		cprintf(c, undr(node))
		cprintln(c, ":")
		for index := range nodeIndexMap {
			cprintln(c, "  Index '%s':", index)
			for _, shard := range nodeIndexMap[index] {
				var prim string
				if shard.Primary {
					prim = "Primary"
				} else {
					prim = "Replica"
				}
				cprintln(c, "    %d: %s, %s", shard.ID, shard.State, prim)
			}
		}
	}

}

type shard struct {
	ID      int
	Primary bool
	State   string
}

func buildNodeIndexInfo(node string, indices []*es.ShortIndexInfo) (map[string][]shard, error) {
	nodeIndexInfo := make(map[string][]shard)
	for _, idx := range indices {
		shards, err := context.IndexShards(idx.Name)
		if err != nil {
			return nil, err
		}
		var nodeShards []shard
		for _, sh := range shards {
			for _, is := range sh.Shards {
				if is.Node.Name == node {
					shard := shard{
						ID:      sh.ID,
						Primary: is.Routing.Primary,
						State:   is.Routing.State,
					}
					nodeShards = append(nodeShards, shard)
				}
			}
		}
		if len(nodeShards) > 0 {
			nodeIndexInfo[idx.Name] = nodeShards
		}
	}
	return nodeIndexInfo, nil
}

func decomissionNodes(c *ishell.Context) {
	if context == nil {
		errorMsg(c, errNotConnected)
		return
	}

	type decomissionArgs struct {
		documentSelectorData
		Mode  string `long:"selector" description:"Selector" choice:"node" choice:"ip" choice:"host" default:""`
		Clear bool   `long:"clear" description:"Removes routing allocation for a given selector"`
	}

	slct, err := parseDocumentArgsCustom(c.Args, &decomissionArgs{})
	if err != nil {
		errorMsg(c, err.Error())
		return
	}
	selector := slct.(*decomissionArgs)

	if len(selector.Mode) < 1 {
		settings, err := context.GetSettings()
		if err != nil {
			errorMsg(c, err.Error())
			return
		}
		cprintlist(c, "Existing transient routing allocations:")

		transient, ok := settings["transient"].(map[string]interface{})
		if !ok {
			cprintln(c, "None")
			return
		}
		cluster, ok := transient["cluster"].(map[string]interface{})
		if !ok {
			cprintln(c, "None")
			return
		}
		routing, ok := cluster["routing"].(map[string]interface{})
		if !ok {
			cprintln(c, "None")
			return
		}
		allocation, ok := routing["allocation"].(map[string]interface{})
		if !ok {
			cprintln(c, "None")
			return
		}
		excludes, ok := allocation["exclude"].(map[string]interface{})
		if !ok {
			cprintln(c, "None")
		} else {
			for key, value := range excludes {
				cprintlist(c, key, ": ", hbl(value.(string)))
			}
		}
	} else {
		var nodes string
		if selector.Clear {
			nodes = ""
		} else {
			nodes = strings.Join(selector.Args, ",")
			if len(nodes) == 0 {
				errorMsg(c, "Please specify allocation rules")
				return
			}
		}

		err = context.DecomissionNode(selector.Mode, nodes)
		if err != nil {
			errorMsg(c, "Failed to modify cluster allocation: "+err.Error())
		} else {
			cprintln(c, "Ok")
		}
	}
}

func filterSettings(settings map[string]interface{}, prefix string) []string {
	result := make([]string, 0)
	for key := range settings {
		if strings.HasPrefix(key, prefix) {
			result = append(result, settings[key].(string))
		}
	}
	return result
}

// -- presentation functions ---

func print(c *ishell.Context, n *es.NodesStats) {
	for nodeID := range n.Nodes {
		cprintf(c, "\n%s", undr(nodeID))
		cprintln(c, ": %v", nodeStatsToString(n.Nodes[nodeID]))
	}
}

func printEnvironment(c *ishell.Context, nodeEnv map[string]es.NodeEnvironmentInfo) {
	for nodeID := range nodeEnv {
		cprintf(c, "\n%s", undr(nodeID))
		cprintln(c, ": %s", environmentToString(nodeEnv[nodeID]))
	}
}

func environmentToString(n es.NodeEnvironmentInfo) string {
	return pad(fmt.Sprintf("\n"+
		"OS: %s\n"+
		"JVM: %s", osToString(n.OS), jvmVerToString(n.JVM)), 2)
}

func osToString(os *es.OSInfo) string {
	return pad(fmt.Sprintf("\n"+
		"%s %s %s\n"+
		"Allocated CPUs: %d", os.Name, os.Arch, os.Version, os.CPUs), 2)
}

func jvmVerToString(j *es.JVMInfo) string {
	return pad(fmt.Sprintf("\n"+
		"%s %s \n"+
		"Vendor: %s", j.VMName, j.Version, j.VMVendor), 2)
}

func nodeStatsToString(n es.NodeStats) string {
	return pad(fmt.Sprintf("\n"+
		"Name: %s\n"+
		"Transport Address: %s\n"+
		"Host: %s\n"+
		"JVM: %s\n"+
		"Indices: %s\n"+
		"Filesystem: %s",
		n.Name, n.TransportAddress, n.Host, jvmToString(n.JVM), indicesToString(n.Indices), fsToString(n.FS)), 2)
}

func indicesToString(i es.NodeIndicesStats) string {
	return pad(fmt.Sprintf(
		"\nDocuments:\n"+
			"  Count: %d\n"+
			"  Deleted: %d\n"+
			"Store:\n"+
			"  Size: %.1f Mb\n"+
			"  Throttle Time: %d sec",
		i.Docs.Count, i.Docs.Deleted, megabytes(i.Store.Size), i.Store.ThrottleTime/1000), 2)
}

func jvmToString(j es.NodeJvmStats) string {
	return pad(fmt.Sprintf(
		"\nUptime: %s"+
			"\nMemory: %s"+
			"\nThreads: %s",
		fmtTime(j.Uptime/1000), memoryToString(j.Memory), threadsToString(j.Threads)), 2)
}

func threadsToString(t *es.NodeThreadStats) string {
	return pad(fmt.Sprintf(
		"\nCount: %d"+
			"\nPeak Count: %d",
		t.Count, t.PeakCount), 2)
}

func memoryToString(m *es.NodeMemoryStats) string {
	return pad(fmt.Sprintf(
		"\n"+
			"Heap Used: %.1f Mb\n"+
			"Heap Committed: %.1f Mb\n"+
			"Heap Max: %.1f Mb\n"+
			"Non-Heap Used: %.1f Mb\n"+
			"Non-Heap Committed: %.1f Mb",
		megabytes(m.HeapUsed), megabytes(m.HeapCommitted), megabytes(m.HeapMax),
		megabytes(m.NonHeapUsed), megabytes(m.NonHeapCommitted)), 2)
}

func fsToString(f es.FsStats) string {
	var buffer bytes.Buffer
	for i, n := range f.Data {
		buffer.WriteString("\n    " + strconv.FormatInt(int64(i), 10) + ":")
		buffer.WriteString(pad(fsNodeToString(n), 4))
		buffer.WriteString("\n")
	}
	data := buffer.String()
	return pad(fmt.Sprintf(
		"\nTotal:\n"+
			"  Total: %6.1f Mb\n"+
			"  Free: %6.1f Mb\n"+
			"  Available: %6.1f Mb\n"+
			"  Data: %s",
		megabytes(f.Total.Total), megabytes(f.Total.Free), megabytes(f.Total.Available),
		data), 2)
}

func fsNodeToString(f es.FsNodeStats) string {
	return pad(fmt.Sprintf(
		"\n"+
			"Path: %s\n"+
			"Mount: %s\n"+
			"FS Type: %s\n"+
			"Total: %.1f Mb\n"+
			"Free: %.1f Mb\n"+
			"Available: %.1f Mb",
		f.Path, f.Mount, f.Type, megabytes(f.Total), megabytes(f.Free), megabytes(f.Available)), 2)
}

func megabytes(b int64) float64 {
	return float64(b) / 1024.0 / 1024.0
}

func fmtTime(seconds int64) string {
	minutes := seconds / 60
	seconds = seconds % 60
	hours := minutes / 60
	minutes = minutes % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

func pad(s string, num int) string {
	return strings.Replace(s, "\n", "\n"+strings.Repeat(" ", num), -1)
}
