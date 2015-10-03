package KinGo

import (
	"strings"
	"regexp"
	"path"
//	"fmt"
)

type ControllerTree struct {
	//search fix route first
	fixrouters map[string]*ControllerTree

	//if set, failure to match fixrouters search then search wildcard
	wildCards *ControllerTree

	//if set, failure to match wildcard search
	leaves []*leafInfo
}

type leafInfo struct {
	// names of wildcards that lead to this leaf. eg, ["id" "name"] for the wildcard ":id" and ":name"
	wildCards []string

	// if the leaf is regexp
	regExps *regexp.Regexp

	runObject interface{}
}

func newControllerTree() *ControllerTree {
	return &ControllerTree{
		fixrouters: make(map[string]*ControllerTree),
	};
}

// call addseg function
// runObject is controllerInfo or interface
// runObject.(*controllerInfo).handler
/*
	fmt.Println( runObject.(*controllerInfo).handler )
	fmt.Println( runObject.(*controllerInfo).controllerType )
	fmt.Println( runObject.(*controllerInfo).methods )
	fmt.Println( runObject.(*controllerInfo).pattern )
	fmt.Println( runObject.(*controllerInfo).routerType )
	fmt.Println( runObject.(*controllerInfo).runfunction )
 */
func ( ct *ControllerTree ) AddRouter( pattern string, runObject interface{}) {
	var segments []string = ct.splitPath( pattern , "/" );
	ct.addseg( segments , runObject, nil, "")
}

func ( ct *ControllerTree ) addseg( segments []string, route interface{}, _wildCards []string, reg string) {
	if len( segments ) > 0  {
		seg := segments[0]
		isWild, params, regexpStr := ct.splitSegment( seg );
//		fmt.Println( isWild );
//		fmt.Println( params );
//		fmt.Println( regexpStr );
		if len( _wildCards ) > 0  {
			if !isWild && inSlice(":splat", _wildCards) {
				isWild = true
				regexpStr = seg
			}
			if seg == "*"  && reg == "" {
				isWild = true
				regexpStr = "(.+)"
			}
		}
		if isWild {
			if ct.wildCards == nil {
				ct.wildCards = newControllerTree()
			}
			if  len( regexpStr ) > 0  {
				if len( reg ) >  0 {
					regexpStr = "/" + regexpStr
				} else {
					var tmpExp  string = "";
					for _ , expReg := range _wildCards {
						if expReg == "." || expReg == ":" {
							continue
						}
						if expReg == ":splat" {
							tmpExp  = tmpExp + "(.+)/"
						} else {
							tmpExp = tmpExp + "([^/]+)/"
						}
					}
					regexpStr = tmpExp + regexpStr
				}
			}else if len(reg) > 0  {
				for _ , expReg := range params {
					if expReg == "." || expReg == ":" {
						continue
					}
					regexpStr = "/([^/]+)" + regexpStr
				}
			}
			ct.wildCards.addseg( segments[1:], route, append( _wildCards , params... ), reg+regexpStr );
		}else{
			subTree, ok := ct.fixrouters[seg]
			if !ok {
				subTree = newControllerTree()
				ct.fixrouters[seg] = subTree
			}
			subTree.addseg(segments[1:], route, _wildCards, reg)
		}
	} else {
		if reg != "" {
			filterCards := []string{}
			for _, v := range _wildCards {
				if v == ":" || v == "." {
					continue
				}
				filterCards = append(filterCards, v)
			}
			ct.leaves = append( ct.leaves, &leafInfo{runObject: route, wildCards: filterCards, regExps: regexp.MustCompile("^" + reg + "$")} );
		} else {
			ct.leaves = append( ct.leaves, &leafInfo{runObject: route, wildCards: _wildCards } );
		}
	}
}


// match router to runObject & params
func ( ct *ControllerTree) Match( pattern string ) ( runObject interface{}, params map[string]string) {
	if len(pattern) == 0 || pattern[0] != '/' {
		return nil, nil
	}

	return ct.match( ct.splitPath(pattern,"/"), nil)
}

func ( ct *ControllerTree) match(segments []string, wildcardValues []string) (runObject interface{}, params map[string]string) {
	// Handle leaf nodes:
	if len(segments) == 0 {
		for _, l := range ct.leaves {
			if ok, pa := l.match(wildcardValues); ok {
				return l.runObject, pa
			}
		}
		if ct.wildCards != nil {
			for _, l := range ct.wildCards.leaves {
				if ok, pa := l.match(wildcardValues); ok {
					return l.runObject, pa
				}
			}

		}
		return nil, nil
	}

	seg, segs := segments[0], segments[1:]

	subTree, ok := ct.fixrouters[seg]
	if ok {
		runObject, params = subTree.match(segs, wildcardValues)
	} else if len(segs) == 0 { //.json .xml
		if subindex := strings.LastIndex(seg, "."); subindex != -1 {
			subTree, ok = ct.fixrouters[seg[:subindex]]
			if ok {
				runObject, params = subTree.match(segs, wildcardValues)
				if runObject != nil {
					if params == nil {
						params = make(map[string]string)
					}
					params[":ext"] = seg[subindex+1:]
					return runObject, params
				}
			}
		}
	}
	if runObject == nil && ct.wildCards != nil {
		runObject, params = ct.wildCards.match(segs, append(wildcardValues, seg))
	}
	if runObject == nil {
		for _, l := range ct.leaves {
			if ok, pa := l.match(append(wildcardValues, segments...)); ok {
				return l.runObject, pa
			}
		}
	}
	return runObject, params
}

func ( ct *ControllerTree)  splitPath( key ,delimiter string ) []string {
	elements := strings.Split(key, delimiter)
	if elements[0] == "" {
		elements = elements[1:]
	}
	if elements[len(elements)-1] == "" {
		elements = elements[:len(elements)-1]
	}
	return elements
}
func inSlice(v string, sl []string) bool {
	for _, val := range sl {
		if val == v {
			return true
		}
	}
	return false
}

// "admin" -> false, nil, ""
// ":id" -> true, [:id], ""
// "?:id" -> true, [: :id], ""        : meaning can empty
// ":id:int" -> true, [:id], ([0-9]+)
// ":name:string" -> true, [:name], ([\w]+)
// ":id([0-9]+)" -> true, [:id], ([0-9]+)
// ":id([0-9]+)_:name" -> true, [:id :name], ([0-9]+)_(.+)
// "cms_:id_:page.html" -> true, [:id :page], cms_(.+)_(.+).html
// "*" -> true, [:splat], ""
// "*.*" -> true,[. :path :ext], ""      . meaning separator
func ( ct *ControllerTree)  splitSegment ( key string ) (bool, []string, string) {
	if strings.HasPrefix(key, "*") {
		if key == "*.*" {
			return true, []string{".", ":path", ":ext"}, ""
		} else {
			return true, []string{":splat"}, ""
		}
	}else if strings.ContainsAny( key , ":") {
		var (
			paramsNum int;
			out []rune;
			start bool;
			startExp bool;
			param []rune;
			expt []rune;
			skipNum int;
			params []string = []string{}
			reg  *regexp.Regexp = regexp.MustCompile(`[a-zA-Z0-9_]+`)
		)
		//便利字符数组
		for index , val := range key {
			if skipNum > 0 {
				skipNum -= 1
				continue
			}
			if start {
				//:id:int and :name:string
				if val == ':' {
					if len( key ) >= index + 4 && key[ index + 1:index + 4] == "int"{
							out = append(out, []rune("([0-9]+)")...);
							params = append(params, ":"+string(param));
							start = false;
							startExp = false;
							skipNum = 3;
							param = make([]rune, 0);
							paramsNum += 1;
							continue;
					}else if len( key ) >= index + 7  && key[index + 1:index + 7] == "string" {
							out = append(out, []rune(`([\w]+)`)...);
							params = append(params, ":"+string(param));
							paramsNum += 1;
							start = false;
							startExp = false;
							skipNum = 6;
							param = make([]rune, 0);
							continue;
					}
				}
				// params only support a-zA-Z0-9
				if reg.MatchString(string(val)) {
					param = append(param, val);
					continue;
				}else if val != '(' {
					out = append(out, []rune(`(.+)`)...);
					params = append(params, ":"+string(param));
					param = make([]rune, 0);
					paramsNum += 1;
					start = false;
					startExp = false;
				}
			}
			if startExp && val != ')' {
					expt = append(expt, val);
					continue;
			}else if val == ':' {
				param = make([]rune, 0);
				start = true;
			}else if val == '(' {
				startExp = true;
				start = false;
				params = append(params, ":"+string(param));
				paramsNum += 1;
				expt = make([]rune, 0);
				expt = append(expt, '(');
			}else if val == ')' {
				startExp = false;
				expt = append(expt, ')');
				out = append(out, expt...);
				param = make([]rune, 0);
			} else if val  == '?' {
				params = append(params, ":");
			} else {
				out = append(out, val);
			}
		}
		if len(param) > 0 {
			if paramsNum > 0 {
				out = append(out, []rune(`(.+)`)...);
			}
			params = append(params, ":"+string(param));
		}
		return true, params, string(out)
    }else{
		return false, nil, ""
	}
}
func (leaf *leafInfo) match(wildcardValues []string) (ok bool, params map[string]string) {
	if leaf.regExps == nil {
		// has error
		if len(wildcardValues) == 0 && len(leaf.wildCards) > 0 {
			if inSlice(":", leaf.wildCards) {
				params = make(map[string]string)
				j := 0
				for _, v := range leaf.wildCards {
					if v == ":" {
						continue
					}
					params[v] = ""
					j += 1
				}
				return true, params
			}
			return false, nil
		} else if len(wildcardValues) == 0 { // static path
			return true, nil
		}
		// match *
		if len(leaf.wildCards) == 1 && leaf.wildCards[0] == ":splat" {
			params = make(map[string]string)
			params[":splat"] = path.Join(wildcardValues...)
			return true, params
		}
		// match *.*
		if len(leaf.wildCards) == 3 && leaf.wildCards[0] == "." {
			params = make(map[string]string)
			lastone := wildcardValues[len(wildcardValues)-1]
			strs := strings.SplitN(lastone, ".", 2)
			if len(strs) == 2 {
				params[":ext"] = strs[1]
			} else {
				params[":ext"] = ""
			}
			params[":path"] = path.Join(wildcardValues[:len(wildcardValues)-1]...) + "/" + strs[0]
			return true, params
		}
		// match :id
		params = make(map[string]string)
		j := 0
		for _, v := range leaf.wildCards {
			if v == ":" {
				continue
			}
			if v == "." {
				lastone := wildcardValues[len(wildcardValues)-1]
				strs := strings.SplitN(lastone, ".", 2)
				if len(strs) == 2 {
					params[":ext"] = strs[1]
				} else {
					params[":ext"] = ""
				}
				if len(wildcardValues[j:]) == 1 {
					params[":path"] = strs[0]
				} else {
					params[":path"] = path.Join(wildcardValues[j:]...) + "/" + strs[0]
				}
				return true, params
			}
			params[v] = wildcardValues[j]
			j += 1
		}
		if len(params) != len(wildcardValues) {
			return false, nil
		}
		return true, params
	}
	if !leaf.regExps.MatchString(path.Join(wildcardValues...)) {
		return false, nil
	}
	params = make(map[string]string)
	matches := leaf.regExps.FindStringSubmatch(path.Join(wildcardValues...))
	for i, match := range matches[1:] {
		params[leaf.wildCards[i]] = match
	}
	return true, params
}
