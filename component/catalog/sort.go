package catalog

type ByPath []File

func (f ByPath) Len() int           { return len(f) }
func (f ByPath) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
func (f ByPath) Less(i, j int) bool { return f[i].Path < f[j].Path }
