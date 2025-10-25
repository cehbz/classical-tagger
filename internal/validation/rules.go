package validation

import (
	"reflect"
	
	"github.com/cehbz/classical-tagger/internal/domain"
)

// RuleFunc is the function signature for all validation rules
type RuleFunc func(actual, reference *domain.Album) RuleResult

// Rules is a collection of validation rules.
// Any exported method on this struct that matches the RuleFunc signature
// is automatically discovered as a validation rule.
type Rules struct{}

// NewRules creates a new Rules instance
func NewRules() *Rules {
	return &Rules{}
}

// Discover finds all rule methods using reflection.
// It looks for exported methods with signature:
//   func (r *Rules) MethodName(actual, reference *domain.Album) RuleResult
func (r *Rules) Discover() []RuleFunc {
	var rules []RuleFunc
	
	rulesType := reflect.TypeOf(r)
	rulesValue := reflect.ValueOf(r)
	
	// Get the Album pointer type for comparison
	albumPtrType := reflect.TypeOf((*domain.Album)(nil))
	
	// Get the RuleResult type for comparison
	ruleResultType := reflect.TypeOf(RuleResult{})
	
	for i := 0; i < rulesType.NumMethod(); i++ {
		method := rulesType.Method(i)
		
		// Skip the Discover method itself
		if method.Name == "Discover" {
			continue
		}
		
		methodType := method.Type
		
		// Check signature:
		// - Must have 3 inputs: receiver (*Rules), actual (*domain.Album), reference (*domain.Album)
		// - Must have 1 output: RuleResult
		if methodType.NumIn() != 3 || methodType.NumOut() != 1 {
			continue
		}
		
		// Check parameter types
		// methodType.In(0) is the receiver (*Rules)
		// methodType.In(1) should be *domain.Album (actual)
		// methodType.In(2) should be *domain.Album (reference)
		if methodType.In(1) != albumPtrType || methodType.In(2) != albumPtrType {
			continue
		}
		
		// Check return type
		if methodType.Out(0) != ruleResultType {
			continue
		}
		
		// This is a valid rule method - wrap it as a RuleFunc
		methodValue := rulesValue.Method(i)
		
		rule := func(actual, reference *domain.Album) RuleResult {
			results := methodValue.Call([]reflect.Value{
				reflect.ValueOf(actual),
				reflect.ValueOf(reference),
			})
			return results[0].Interface().(RuleResult)
		}
		
		rules = append(rules, rule)
	}
	
	return rules
}

// All is a convenience function that creates a Rules instance and discovers all rules
func All() []RuleFunc {
	return NewRules().Discover()
}