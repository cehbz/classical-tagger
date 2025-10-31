package validation

import (
	"reflect"

	"github.com/cehbz/classical-tagger/internal/domain"
)

// AlbumRuleFunc is the signature for album-level validation rules
type AlbumRuleFunc func(actual, reference *domain.Album) RuleResult

// TrackRuleFunc is the signature for track-level validation rules
type TrackRuleFunc func(actualTrack, refTrack *domain.Track, actualAlbum, refAlbum *domain.Album) RuleResult

// Rules is a collection of validation rules.
// Any exported method on this struct that matches either AlbumRuleFunc or TrackRuleFunc signature
// is automatically discovered as a validation rule.
type Rules struct{}

// NewRules creates a new Rules instance
func NewRules() *Rules {
	return &Rules{}
}

// AlbumRules finds all album-level rule methods using reflection.
// It looks for exported methods with signature:
//
//	func (r *Rules) MethodName(actual, reference *domain.Album) RuleResult
func (r *Rules) AlbumRules() []AlbumRuleFunc {
	var rules []AlbumRuleFunc

	rulesType := reflect.TypeOf(r)
	rulesValue := reflect.ValueOf(r)

	// Get the Album pointer type for comparison
	albumPtrType := reflect.TypeOf((*domain.Album)(nil))

	// Get the RuleResult type for comparison
	ruleResultType := reflect.TypeOf(RuleResult{})

	for i := 0; i < rulesType.NumMethod(); i++ {
		method := rulesType.Method(i)

		// Skip discovery methods themselves
		if method.Name == "AlbumRules" || method.Name == "TrackRules" {
			continue
		}

		methodType := method.Type

		// Check signature for album-level rule:
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

		// This is a valid album-level rule method - wrap it as an AlbumRuleFunc
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

// TrackRules finds all track-level rule methods using reflection.
// It looks for exported methods with signature:
//
//	func (r *Rules) MethodName(actualTrack, refTrack *domain.Track, actualAlbum, refAlbum *domain.Album) RuleResult
func (r *Rules) TrackRules() []TrackRuleFunc {
	var rules []TrackRuleFunc

	rulesType := reflect.TypeOf(r)
	rulesValue := reflect.ValueOf(r)

	// Get the Track pointer type for comparison
	trackPtrType := reflect.TypeOf((*domain.Track)(nil))

	// Get the Album pointer type for comparison
	albumPtrType := reflect.TypeOf((*domain.Album)(nil))

	// Get the RuleResult type for comparison
	ruleResultType := reflect.TypeOf(RuleResult{})

	for i := 0; i < rulesType.NumMethod(); i++ {
		method := rulesType.Method(i)

		// Skip discovery methods themselves
		if method.Name == "AlbumRules" || method.Name == "TrackRules" {
			continue
		}

		methodType := method.Type

		// Check signature for track-level rule:
		// - Must have 5 inputs: receiver (*Rules), actualTrack (*Track), refTrack (*Track), actualAlbum (*Album), refAlbum (*Album)
		// - Must have 1 output: RuleResult
		if methodType.NumIn() != 5 || methodType.NumOut() != 1 {
			continue
		}

		// Check parameter types
		// methodType.In(0) is the receiver (*Rules)
		// methodType.In(1) should be *domain.Track (actualTrack)
		// methodType.In(2) should be *domain.Track (refTrack)
		// methodType.In(3) should be *domain.Album (actualAlbum)
		// methodType.In(4) should be *domain.Album (refAlbum)
		if methodType.In(1) != trackPtrType || methodType.In(2) != trackPtrType ||
			methodType.In(3) != albumPtrType || methodType.In(4) != albumPtrType {
			continue
		}

		// Check return type
		if methodType.Out(0) != ruleResultType {
			continue
		}

		// This is a valid track-level rule method - wrap it as a TrackRuleFunc
		methodValue := rulesValue.Method(i)

		rule := func(actualTrack, refTrack *domain.Track, actualAlbum, refAlbum *domain.Album) RuleResult {
			results := methodValue.Call([]reflect.Value{
				reflect.ValueOf(actualTrack),
				reflect.ValueOf(refTrack),
				reflect.ValueOf(actualAlbum),
				reflect.ValueOf(refAlbum),
			})
			return results[0].Interface().(RuleResult)
		}

		rules = append(rules, rule)
	}

	return rules
}
