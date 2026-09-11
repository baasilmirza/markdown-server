package main

import (
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// doc describes a single markdown file in the catalog.
type doc struct {
	Rel   string // slash-separated path relative to root, e.g. "guide/advanced.md"
	Route string // route without the .md extension, e.g. "guide/advanced"
	URL   string // escaped href for the route
	Title string // display name, base name without extension
}

// node is a directory or file in the recursive tree.
type node struct {
	Name     string
	Route    string
	URL      string
	IsDir    bool
	Children []*node
}

// catalog is a snapshot of the markdown tree rooted at root.
type catalog struct {
	root  string
	tree  []*node
	docs  []doc
	byRel map[string]doc
}

// buildCatalog walks root and returns a catalog of every markdown file.
func buildCatalog(root string) (*catalog, error) {
	tree, err := walkTree(root, "")
	if err != nil {
		return nil, err
	}
	c := &catalog{root: root, tree: tree, byRel: make(map[string]doc)}
	c.flatten()
	return c, nil
}

func walkTree(root, rel string) ([]*node, error) {
	dir := filepath.Join(root, filepath.FromSlash(rel))
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var nodes []*node
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		childRel := name
		if rel != "" {
			childRel = rel + "/" + name
		}
		if e.IsDir() {
			children, err := walkTree(root, childRel)
			if err != nil {
				continue
			}
			nodes = append(nodes, &node{Name: name, IsDir: true, Children: children})
			continue
		}
		if !strings.EqualFold(filepath.Ext(name), ".md") {
			continue
		}
		route := strings.TrimSuffix(childRel, filepath.Ext(childRel))
		nodes = append(nodes, &node{
			Name:  name,
			Route: route,
			URL:   routeURL(route),
		})
	}
	sortNodes(nodes)
	return nodes, nil
}

// sortNodes orders directories before files, then alphabetically.
func sortNodes(nodes []*node) {
	sort.SliceStable(nodes, func(i, j int) bool {
		if nodes[i].IsDir != nodes[j].IsDir {
			return nodes[i].IsDir
		}
		return strings.ToLower(nodes[i].Name) < strings.ToLower(nodes[j].Name)
	})
}

func (c *catalog) flatten() {
	var walk func(nodes []*node)
	walk = func(nodes []*node) {
		for _, n := range nodes {
			if n.IsDir {
				walk(n.Children)
				continue
			}
			title := strings.TrimSuffix(n.Name, filepath.Ext(n.Name))
			d := doc{
				Rel:   n.Route + ".md",
				Route: n.Route,
				URL:   n.URL,
				Title: title,
			}
			c.docs = append(c.docs, d)
			c.byRel[d.Rel] = d
		}
	}
	walk(c.tree)
}

// sidebarItem is a flattened row for rendering the tree.
type sidebarItem struct {
	Name   string
	Route  string
	URL    string
	Depth  int
	Indent int
	IsDir  bool
	Active bool
}

// sidebar flattens the tree into rows, marking the row matching route active.
func (c *catalog) sidebar(route string) []sidebarItem {
	var items []sidebarItem
	var walk func(nodes []*node, depth int)
	walk = func(nodes []*node, depth int) {
		for _, n := range nodes {
			item := sidebarItem{
				Name:   n.Name,
				Route:  n.Route,
				URL:    n.URL,
				Depth:  depth,
				Indent: depth*14 + 8,
				IsDir:  n.IsDir,
			}
			if n.IsDir {
				items = append(items, item)
				walk(n.Children, depth+1)
				continue
			}
			item.Name = strings.TrimSuffix(n.Name, filepath.Ext(n.Name))
			item.Active = route != "" && n.Route == route
			items = append(items, item)
		}
	}
	walk(c.tree, 0)
	return items
}

// lookup returns the doc for a slash-separated relative path.
func (c *catalog) lookup(rel string) (doc, bool) {
	d, ok := c.byRel[filepath.ToSlash(rel)]
	return d, ok
}

// neighbors returns the documents before and after route in flat order.
func (c *catalog) neighbors(route string) (prev, next *doc) {
	for i, d := range c.docs {
		if d.Route != route {
			continue
		}
		if i > 0 {
			p := c.docs[i-1]
			prev = &p
		}
		if i+1 < len(c.docs) {
			n := c.docs[i+1]
			next = &n
		}
		return
	}
	return
}

// routeURL escapes each path segment of a route for use in an href.
func routeURL(route string) string {
	if route == "" {
		return "/"
	}
	parts := strings.Split(route, "/")
	for i, p := range parts {
		parts[i] = url.PathEscape(p)
	}
	return "/" + strings.Join(parts, "/")
}
