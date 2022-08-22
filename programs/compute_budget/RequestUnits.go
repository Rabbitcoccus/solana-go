// Code generated by https://github.com/gagliardetto/anchor-go. DO NOT EDIT.

package compute_budget

import (
	"errors"
	ag_format "github.com/desperatee/solana-go/text/format"
	ag_binary "github.com/gagliardetto/binary"
	ag_treeout "github.com/gagliardetto/treeout"
)

// RequestUnits is the `requestUnits` instruction.
type RequestUnits struct {
	Units         *uint32
	AdditionalFee *uint32
}

// NewRequestUnitsInstructionBuilder creates a new `RequestUnits` instruction builder.
func NewRequestUnitsInstructionBuilder() *RequestUnits {
	nd := &RequestUnits{}
	return nd
}

// SetUnits sets the "units" parameter.
func (inst *RequestUnits) SetUnits(units uint32) *RequestUnits {
	inst.Units = &units
	return inst
}

// SetAdditionalFee sets the "additionalFee" parameter.
func (inst *RequestUnits) SetAdditionalFee(additionalFee uint32) *RequestUnits {
	inst.AdditionalFee = &additionalFee
	return inst
}

func (inst RequestUnits) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: ag_binary.NoTypeIDDefaultID,
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst RequestUnits) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *RequestUnits) Validate() error {
	// Check whether all (required) parameters are set:
	{
		if inst.Units == nil {
			return errors.New("Units parameter is not set")
		}
		if inst.AdditionalFee == nil {
			return errors.New("AdditionalFee parameter is not set")
		}
	}

	// Check whether all (required) accounts are set:
	{
	}
	return nil
}

func (inst *RequestUnits) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("RequestUnits")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params[len=2]").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param("        Units", *inst.Units))
						paramsBranch.Child(ag_format.Param("AdditionalFee", *inst.AdditionalFee))
					})
				})
		})
}

func (obj RequestUnits) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `Units` param:
	err = encoder.Encode(obj.Units)
	if err != nil {
		return err
	}
	// Serialize `AdditionalFee` param:
	err = encoder.Encode(obj.AdditionalFee)
	if err != nil {
		return err
	}
	return nil
}
func (obj *RequestUnits) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `Units`:
	err = decoder.Decode(&obj.Units)
	if err != nil {
		return err
	}
	// Deserialize `AdditionalFee`:
	err = decoder.Decode(&obj.AdditionalFee)
	if err != nil {
		return err
	}
	return nil
}

// NewRequestUnitsInstruction declares a new RequestUnits instruction with the provided parameters and accounts.
func NewRequestUnitsInstruction(
	// Parameters:
	units uint32,
	additionalFee uint32) *RequestUnits {
	return NewRequestUnitsInstructionBuilder().
		SetUnits(units).
		SetAdditionalFee(additionalFee)
}
