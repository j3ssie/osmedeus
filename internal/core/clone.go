package core

// Clone creates a deep copy of the Workflow
func (w *Workflow) Clone() *Workflow {
	if w == nil {
		return nil
	}

	cloned := &Workflow{
		Kind:         w.Kind,
		Name:         w.Name,
		Description:  w.Description,
		Runner:       w.Runner,
		FilePath:     w.FilePath,
		Checksum:     w.Checksum,
		ResolvedFrom: w.ResolvedFrom,
		Extends:      w.Extends,
	}

	// Deep copy Tags
	if len(w.Tags) > 0 {
		cloned.Tags = make(TagList, len(w.Tags))
		copy(cloned.Tags, w.Tags)
	}

	// Deep copy Params
	if len(w.Params) > 0 {
		cloned.Params = make([]Param, len(w.Params))
		copy(cloned.Params, w.Params)
	}

	// Deep copy Triggers
	if len(w.Triggers) > 0 {
		cloned.Triggers = make([]Trigger, len(w.Triggers))
		for i, t := range w.Triggers {
			cloned.Triggers[i] = *t.Clone()
		}
	}

	// Deep copy Dependencies
	cloned.Dependencies = w.Dependencies.Clone()

	// Deep copy Reports
	if len(w.Reports) > 0 {
		cloned.Reports = make([]Report, len(w.Reports))
		copy(cloned.Reports, w.Reports)
	}

	// Deep copy Preferences
	cloned.Preferences = w.Preferences.Clone()

	// Deep copy RunnerConfig
	cloned.RunnerConfig = w.RunnerConfig.Clone()

	// Deep copy Steps (module-specific)
	if len(w.Steps) > 0 {
		cloned.Steps = make([]Step, len(w.Steps))
		for i, s := range w.Steps {
			cloned.Steps[i] = *s.Clone()
		}
	}

	// Deep copy Modules (flow-specific)
	if len(w.Modules) > 0 {
		cloned.Modules = make([]ModuleRef, len(w.Modules))
		for i, m := range w.Modules {
			cloned.Modules[i] = *m.Clone()
		}
	}

	// Deep copy Override (if present)
	cloned.Override = w.Override.Clone()

	return cloned
}

// Clone creates a deep copy of Dependencies
func (d *Dependencies) Clone() *Dependencies {
	if d == nil {
		return nil
	}

	cloned := &Dependencies{}

	if len(d.Commands) > 0 {
		cloned.Commands = make([]string, len(d.Commands))
		copy(cloned.Commands, d.Commands)
	}

	if len(d.Files) > 0 {
		cloned.Files = make([]string, len(d.Files))
		copy(cloned.Files, d.Files)
	}

	if len(d.Variables) > 0 {
		cloned.Variables = make([]VariableDep, len(d.Variables))
		copy(cloned.Variables, d.Variables)
	}

	if len(d.TargetTypes) > 0 {
		cloned.TargetTypes = make([]TargetType, len(d.TargetTypes))
		copy(cloned.TargetTypes, d.TargetTypes)
	}

	if len(d.FunctionsConditions) > 0 {
		cloned.FunctionsConditions = make([]string, len(d.FunctionsConditions))
		copy(cloned.FunctionsConditions, d.FunctionsConditions)
	}

	return cloned
}

// Clone creates a deep copy of Preferences
func (p *Preferences) Clone() *Preferences {
	if p == nil {
		return nil
	}

	cloned := &Preferences{}

	if p.DisableNotifications != nil {
		v := *p.DisableNotifications
		cloned.DisableNotifications = &v
	}
	if p.DisableLogging != nil {
		v := *p.DisableLogging
		cloned.DisableLogging = &v
	}
	if p.HeuristicsCheck != nil {
		v := *p.HeuristicsCheck
		cloned.HeuristicsCheck = &v
	}
	if p.CIOutputFormat != nil {
		v := *p.CIOutputFormat
		cloned.CIOutputFormat = &v
	}
	if p.Silent != nil {
		v := *p.Silent
		cloned.Silent = &v
	}
	if p.Repeat != nil {
		v := *p.Repeat
		cloned.Repeat = &v
	}
	if p.RepeatWaitTime != nil {
		v := *p.RepeatWaitTime
		cloned.RepeatWaitTime = &v
	}

	return cloned
}

// Clone creates a deep copy of RunnerConfig
func (r *RunnerConfig) Clone() *RunnerConfig {
	if r == nil {
		return nil
	}

	cloned := &RunnerConfig{
		Image:      r.Image,
		Network:    r.Network,
		Persistent: r.Persistent,
		Host:       r.Host,
		Port:       r.Port,
		User:       r.User,
		KeyFile:    r.KeyFile,
		Password:   r.Password,
		WorkDir:    r.WorkDir,
	}

	if len(r.Env) > 0 {
		cloned.Env = make(map[string]string, len(r.Env))
		for k, v := range r.Env {
			cloned.Env[k] = v
		}
	}

	if len(r.Volumes) > 0 {
		cloned.Volumes = make([]string, len(r.Volumes))
		copy(cloned.Volumes, r.Volumes)
	}

	return cloned
}

// Clone creates a deep copy of Trigger
func (t *Trigger) Clone() *Trigger {
	if t == nil {
		return nil
	}

	cloned := &Trigger{
		Name:     t.Name,
		On:       t.On,
		Schedule: t.Schedule,
		Path:     t.Path,
		Enabled:  t.Enabled,
		Input:    t.Input, // TriggerInput has no pointer fields, safe to copy
	}

	// Deep copy EventConfig
	if t.Event != nil {
		cloned.Event = &EventConfig{
			Topic: t.Event.Topic,
		}
		if len(t.Event.Filters) > 0 {
			cloned.Event.Filters = make([]string, len(t.Event.Filters))
			copy(cloned.Event.Filters, t.Event.Filters)
		}
	}

	return cloned
}

// Clone creates a deep copy of ModuleRef
func (m *ModuleRef) Clone() *ModuleRef {
	if m == nil {
		return nil
	}

	cloned := &ModuleRef{
		Name:      m.Name,
		Path:      m.Path,
		Condition: m.Condition,
	}

	if len(m.Params) > 0 {
		cloned.Params = make(map[string]string, len(m.Params))
		for k, v := range m.Params {
			cloned.Params[k] = v
		}
	}

	if len(m.DependsOn) > 0 {
		cloned.DependsOn = make([]string, len(m.DependsOn))
		copy(cloned.DependsOn, m.DependsOn)
	}

	if len(m.OnSuccess) > 0 {
		cloned.OnSuccess = make([]Action, len(m.OnSuccess))
		for i, a := range m.OnSuccess {
			cloned.OnSuccess[i] = *a.Clone()
		}
	}

	if len(m.OnError) > 0 {
		cloned.OnError = make([]Action, len(m.OnError))
		for i, a := range m.OnError {
			cloned.OnError[i] = *a.Clone()
		}
	}

	cloned.Decision = m.Decision.Clone()

	return cloned
}

// Clone creates a deep copy of Action
func (a *Action) Clone() *Action {
	if a == nil {
		return nil
	}

	cloned := &Action{
		Action:    a.Action,
		Message:   a.Message,
		Condition: a.Condition,
		Name:      a.Name,
		Value:     a.Value, // interface{} - shallow copy is acceptable
		Type:      a.Type,
		Command:   a.Command,
		Notify:    a.Notify,
	}

	if len(a.Functions) > 0 {
		cloned.Functions = make([]string, len(a.Functions))
		copy(cloned.Functions, a.Functions)
	}

	if len(a.Export) > 0 {
		cloned.Export = make(map[string]string, len(a.Export))
		for k, v := range a.Export {
			cloned.Export[k] = v
		}
	}

	return cloned
}

// Clone creates a deep copy of DecisionConfig
func (d *DecisionConfig) Clone() *DecisionConfig {
	if d == nil {
		return nil
	}

	cloned := &DecisionConfig{
		Switch: d.Switch,
	}

	if len(d.Cases) > 0 {
		cloned.Cases = make(map[string]DecisionCase, len(d.Cases))
		for k, v := range d.Cases {
			cloned.Cases[k] = v
		}
	}

	if d.Default != nil {
		cloned.Default = &DecisionCase{
			Goto: d.Default.Goto,
		}
	}

	return cloned
}

// Clone creates a deep copy of WorkflowOverride
func (o *WorkflowOverride) Clone() *WorkflowOverride {
	if o == nil {
		return nil
	}

	cloned := &WorkflowOverride{}

	// Deep copy Params
	if len(o.Params) > 0 {
		cloned.Params = make(map[string]*ParamOverride, len(o.Params))
		for k, v := range o.Params {
			cloned.Params[k] = v.Clone()
		}
	}

	// Deep copy Steps
	cloned.Steps = o.Steps.Clone()

	// Deep copy Modules
	cloned.Modules = o.Modules.Clone()

	// Deep copy Triggers
	if len(o.Triggers) > 0 {
		cloned.Triggers = make([]Trigger, len(o.Triggers))
		for i, t := range o.Triggers {
			cloned.Triggers[i] = *t.Clone()
		}
	}

	// Deep copy Dependencies
	cloned.Dependencies = o.Dependencies.Clone()

	// Deep copy Preferences
	cloned.Preferences = o.Preferences.Clone()

	// Deep copy RunnerConfig
	cloned.RunnerConfig = o.RunnerConfig.Clone()

	// Deep copy Runner
	if o.Runner != nil {
		v := *o.Runner
		cloned.Runner = &v
	}

	return cloned
}

// Clone creates a deep copy of ParamOverride
func (p *ParamOverride) Clone() *ParamOverride {
	if p == nil {
		return nil
	}

	cloned := &ParamOverride{
		Default: p.Default, // interface{} - shallow copy is acceptable
	}

	if p.Type != nil {
		v := *p.Type
		cloned.Type = &v
	}
	if p.Required != nil {
		v := *p.Required
		cloned.Required = &v
	}
	if p.Generator != nil {
		v := *p.Generator
		cloned.Generator = &v
	}

	return cloned
}

// Clone creates a deep copy of StepsOverride
func (s *StepsOverride) Clone() *StepsOverride {
	if s == nil {
		return nil
	}

	cloned := &StepsOverride{
		Mode: s.Mode,
	}

	if len(s.Steps) > 0 {
		cloned.Steps = make([]Step, len(s.Steps))
		for i, step := range s.Steps {
			cloned.Steps[i] = *step.Clone()
		}
	}

	if len(s.Remove) > 0 {
		cloned.Remove = make([]string, len(s.Remove))
		copy(cloned.Remove, s.Remove)
	}

	if len(s.Replace) > 0 {
		cloned.Replace = make([]Step, len(s.Replace))
		for i, step := range s.Replace {
			cloned.Replace[i] = *step.Clone()
		}
	}

	return cloned
}

// Clone creates a deep copy of ModulesOverride
func (m *ModulesOverride) Clone() *ModulesOverride {
	if m == nil {
		return nil
	}

	cloned := &ModulesOverride{
		Mode: m.Mode,
	}

	if len(m.Modules) > 0 {
		cloned.Modules = make([]ModuleRef, len(m.Modules))
		for i, mod := range m.Modules {
			cloned.Modules[i] = *mod.Clone()
		}
	}

	if len(m.Remove) > 0 {
		cloned.Remove = make([]string, len(m.Remove))
		copy(cloned.Remove, m.Remove)
	}

	if len(m.Replace) > 0 {
		cloned.Replace = make([]ModuleRef, len(m.Replace))
		for i, mod := range m.Replace {
			cloned.Replace[i] = *mod.Clone()
		}
	}

	return cloned
}
