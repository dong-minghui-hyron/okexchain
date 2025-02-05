package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/okex/okexchain/x/gov/types"
	"strings"
)

const (
	// proposalTypeManageWhiteList defines the type for a ManageWhiteListProposal
	proposalTypeManageWhiteList = "ManageWhiteList"
)

func init() {
	govtypes.RegisterProposalType(proposalTypeManageWhiteList)
	govtypes.RegisterProposalTypeCodec(ManageWhiteListProposal{}, "okexchain/farm/ManageWhiteListProposal")
}

var _ govtypes.Content = (*ManageWhiteListProposal)(nil)

// ManageWhiteListProposal - structure for the proposal to add or delete a pool name from white list
type ManageWhiteListProposal struct {
	Title       string `json:"title" yaml:"title"`
	Description string `json:"description" yaml:"description"`
	PoolName    string `json:"pool_name" yaml:"pool_name"`
	IsAdded     bool   `json:"is_added" yaml:"is_added"`
}

// NewManageWhiteListProposal creates a new instance of ManageWhiteListProposal
func NewManageWhiteListProposal(title, description, poolName string, isAdded bool) ManageWhiteListProposal {
	return ManageWhiteListProposal{
		Title:       title,
		Description: description,
		PoolName:    poolName,
		IsAdded:     isAdded,
	}
}

// GetTitle returns title of a manage white list proposal object
func (mp ManageWhiteListProposal) GetTitle() string {
	return mp.Title
}

// GetDescription returns description of a manage white list proposal object
func (mp ManageWhiteListProposal) GetDescription() string {
	return mp.Description
}

// ProposalRoute returns route key of a manage white list proposal object
func (mp ManageWhiteListProposal) ProposalRoute() string {
	return RouterKey
}

// ProposalType returns type of a manage white list proposal object
func (mp ManageWhiteListProposal) ProposalType() string {
	return proposalTypeManageWhiteList
}

// ValidateBasic validates a manage white list proposal
func (mp ManageWhiteListProposal) ValidateBasic() sdk.Error {
	if len(strings.TrimSpace(mp.Title)) == 0 {
		return govtypes.ErrInvalidProposalContent(
			DefaultCodespace,
			"failed to submit the manage white list proposal because the title is blank")
	}
	if len(mp.Title) > govtypes.MaxTitleLength {
		return govtypes.ErrInvalidProposalContent(
			DefaultCodespace,
			fmt.Sprintf("failed to submit the manage white list proposal because the title is longer than max length of %d",
				govtypes.MaxTitleLength))
	}

	if len(mp.Description) == 0 {
		return govtypes.ErrInvalidProposalContent(
			DefaultCodespace,
			"failed to submit the manage white list proposal because the description is blank")
	}

	if len(mp.Description) > govtypes.MaxDescriptionLength {
		return govtypes.ErrInvalidProposalContent(
			DefaultCodespace,
			fmt.Sprintf("failed to submit the manage white list proposal because the description is longer than max length of %d",
				govtypes.MaxDescriptionLength))
	}

	if mp.ProposalType() != proposalTypeManageWhiteList {
		return govtypes.ErrInvalidProposalType(DefaultCodespace, mp.ProposalType())
	}

	if len(mp.PoolName) == 0 {
		return govtypes.ErrInvalidProposalContent(
			DefaultCodespace,
			"failed to submit the manage white list proposal because of the empty target pool name",
		)
	}

	return nil
}

// String returns a human readable string representation of a ManageWhiteListProposal
func (mp ManageWhiteListProposal) String() string {
	return fmt.Sprintf(`ManagerWhiteListProposal:
 Title:					%s
 Description:        	%s
 Type:                	%s
 PoolName:				%s
 IsAdded:				%t`,
		mp.Title, mp.Description, mp.ProposalType(), mp.PoolName, mp.IsAdded)
}
