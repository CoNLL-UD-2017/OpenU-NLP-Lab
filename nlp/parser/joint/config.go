package joint

import (
	. "chukuparser/alg"
	"chukuparser/alg/search"
	"chukuparser/alg/transition"
	dep "chukuparser/nlp/parser/dependency/transition"
	"chukuparser/nlp/parser/disambig"
	nlp "chukuparser/nlp/types"
	"chukuparser/util"

	"fmt"
)

type JointConfig struct {
	dep.SimpleConfiguration
	disambig.MDConfig

	InternalPrevious *JointConfig
	Last             transition.Transition
	ETrans           *util.EnumSet
}

var (
	_ transition.Configuration    = &JointConfig{}
	_ search.Aligned              = &JointConfig{}
	_ dep.DependencyConfiguration = &JointConfig{}
	_ nlp.DependencyGraph         = &JointConfig{}
	_ nlp.MorphDependencyGraph    = &JointConfig{}
)

func (c *JointConfig) Init(abstractLattice interface{}) {
	// initialize MDConfig as usual (doesn't know the difference)
	c.MDConfig.Init(abstractLattice)

	// initialize SimpleConfiguration explicitly
	// we don't know # of morphemes in advance, only an estimate
	estMorphemes := len(c.MDConfig.Lattices) * 2
	c.SimpleConfiguration.InternalStack = NewStackArray(estMorphemes)
	c.SimpleConfiguration.InternalQueue = NewQueueSlice(estMorphemes)
	c.SimpleConfiguration.InternalArcs = dep.NewArcSetSimple(estMorphemes)
	c.SimpleConfiguration.NumHeadStack = 0
	c.SimpleConfiguration.Nodes = make([]*dep.ArcCachedDepNode, 0, estMorphemes)

	// note we don't initialize the queue at all, morph. disambig. will enqueue

	c.Last = 0
	c.InternalPrevious = nil
}

func (c *JointConfig) Terminal() bool {
	return c.MDConfig.Terminal() && c.SimpleConfiguration.Terminal()
}

func (c *JointConfig) Copy() transition.Configuration {
	newConf := new(JointConfig)
	c.CopyTo(newConf)
	return newConf
}

func (c *JointConfig) CopyTo(target transition.Configuration) {
	newConf, ok := target.(*JointConfig)
	if !ok {
		panic("Can't copy into non *JointConfig")
	}

	newConf.Last = c.Last
	newConf.InternalPrevious = c
	newConf.ETrans = c.ETrans
	c.MDConfig.CopyTo(&newConf.MDConfig)
	c.SimpleConfiguration.CopyTo(&newConf.SimpleConfiguration)
}

func (c *JointConfig) GetSequence() transition.ConfigurationSequence {
	if c.Mappings == nil || c.Arcs() == nil {
		return make(transition.ConfigurationSequence, 0)
	}
	retval := make(transition.ConfigurationSequence, 0, len(c.Morphemes)+c.Arcs().Size())
	currentConf := c
	for currentConf != nil {
		retval = append(retval, currentConf)
		currentConf = currentConf.InternalPrevious
	}
	return retval

}

func (c *JointConfig) SetLastTransition(t transition.Transition) {
	c.Last = t
}

func (c *JointConfig) GetLastTransition() transition.Transition {
	return c.Last
}

func (c *JointConfig) String() string {
	if c.Mappings == nil {
		return fmt.Sprintf("\t=>\t([],\t[],\t[],\t,\t)")
	}
	var trans string
	if c.Last < 0 {
		trans = ""
	} else {
		trans = c.SimpleConfiguration.ETrans.ValueOf(int(c.Last)).(string)
	}
	mapLen := len(c.Mappings)
	if mapLen > 0 {
		return fmt.Sprintf("%s\t=>\t([%s],\t[%s],\t[%s],\t[%s],\t[%v])",
			trans,
			c.StringStack(),
			c.StringQueue(),
			c.StringLatticeQueue(),
			c.StringArcs(),
			c.Mappings[mapLen-1])
	} else {
		return fmt.Sprintf("%s\t=>\t([],\t[],\t[%s],\t,\t)", trans, c.StringLatticeQueue())
	}
}

func (c *JointConfig) Equal(otherEq util.Equaler) bool {
	if (otherEq == nil && c != nil) || (c == nil && otherEq != nil) {
		return false
	}
	switch other := otherEq.(type) {
	case *JointConfig:
		if (other == nil && c != nil) || (c == nil && other != nil) {
			return false
		}
		if other.Last != c.Last {
			return false
		}
		if c.InternalPrevious == nil && other.InternalPrevious == nil {
			return true
		}
		if c.InternalPrevious != nil && other.InternalPrevious != nil {
			return c.InternalPrevious.Equal(other.InternalPrevious)
		} else {
			return false
		}
	}
	panic("Can't equal to non-Joint config")
}

func (c *JointConfig) Address(location []byte, offset int) (nodeID int, exists bool, isGenerator bool) {
	if location[0] == 'M' || location[0] == 'L' {
		return c.MDConfig.Address(location, offset)
	} else {
		return c.SimpleConfiguration.Address(location, offset)
	}
}

func (c *JointConfig) GenerateAddresses(nodeID int, location []byte) (nodeIDs []int) {
	return c.SimpleConfiguration.GenerateAddresses(nodeID, location)
}

func (c *JointConfig) Attribute(source byte, nodeID int, attribute []byte) (interface{}, bool) {
	if source == 'M' || source == 'L' {
		return c.MDConfig.Attribute(source, nodeID, attribute)
	} else {
		return c.SimpleConfiguration.Attribute(source, nodeID, attribute)
	}
}

func (c *JointConfig) Previous() transition.Configuration {
	return c.InternalPrevious
}

func (c *JointConfig) Clear() {
	c.SimpleConfiguration.Clear()
	c.MDConfig.Clear()
	c.InternalPrevious = nil
}

func (c *JointConfig) Alignment() int {
	return len(c.MDConfig.Mappings)
}

func (c *JointConfig) GetMappings() nlp.Mappings {
	return c.Mappings
}

func (c *JointConfig) GetMorpheme(i int) *nlp.EMorpheme {
	return c.Morphemes[i]
}