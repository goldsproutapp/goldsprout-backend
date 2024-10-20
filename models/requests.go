package models

type StockSnapshotCreationPayload struct {
	StockName      string `binding:"required" json:"stock_name"`
	StockCode      string `binding:"required" json:"stock_code"`
	Units          string `binding:"required" json:"units"`
	Price          string `binding:"required" json:"price"`
	Cost           string `binding:"required" json:"cost"`
	Value          string `binding:"required" json:"value"`
	AbsoluteChange string `binding:"required" json:"absolute_change"`

	Sector    string `json:"sector"`
	Region    string `json:"region"`
	AnnualFee string `json:"annual_fee"`
}

type StockSnapshotCreationBatch struct {
	Entries          []StockSnapshotCreationPayload `binding:"required" json:"entries"`
	AccountID        uint                           `binding:"required" json:"account_id"`
	Date             int64                          `binding:"required" json:"date"`
	DeleteSoldStocks bool                           `json:"delete_sold_stocks"`
}
type StockSnapshotCreationRequest struct {
	Batches []StockSnapshotCreationBatch `json:"batches,omitempty" binding:"required"`
}

type StockUpdateRequest struct {
	Stock Stock `binding:"required" json:"stock"`
}

type ProviderUpdateRequest struct {
	Provider Provider `binding:"required" json:"provider"`
}

type StockFilterQuery struct {
	FilterRegions      string `json:"filter_regions,omitempty" form:"filter_regions"`
	FilterProviders    string `json:"filter_providers,omitempty" form:"filter_providers"`
	FilterUsers        string `json:"filter_users,omitempty" form:"filter_users"`
	FilterAccounts     string `json:"filter_accounts,omitempty" form:"filter_accounts"`
	FilterIgnoreBefore string `json:"filter_ignore_before,omitempty" form:"filter_ignore_before"`
	FilterIgnoreAfter  string `json:"filter_ignore_after,omitempty" form:"filter_ignore_after"`
}

type PerformanceRequestQuery struct {
	StockFilterQuery
	Compare string `binding:"required" json:"compare,omitempty" form:"compare"`
	Of      string `binding:"required" json:"of,omitempty" form:"of"`
	For     string `binding:"required" json:"for,omitempty" form:"for"`
	Over    string `binding:"required" json:"over,omitempty" form:"over"`
}

type SplitRequestQuery struct {
	StockFilterQuery
	Compare string `binding:"required" json:"compare,omitempty" form:"compare"`
	Across  string `binding:"required" json:"across,omitempty" form:"across"`
}

type UserInvitationRequest struct {
	Email     string `binding:"required" json:"email,omitempty"`
	FirstName string `binding:"required" json:"first_name,omitempty"`
	LastName  string `binding:"required" json:"last_name,omitempty"`
}
type UserInvitationAccept struct {
	Token    string `binding:"required" json:"token,omitempty"`
	Password string `binding:"required" json:"password,omitempty"`
}

type PasswordChangeRequest struct {
	OldPassword string `binding:"required" json:"old_password,omitempty"`
	NewPassword string `binding:"required" json:"new_password,omitempty"`
}

type SetPermissionsRequest struct {
	User        uint                        `binding:"required" json:"user,omitempty"`
	Permissions []SetPermissionsRequestItem `binding:"required" json:"permissions,omitempty"`
	Trusted     *bool                       `binding:"required" json:"trusted,omitempty"`
}
type SetPermissionsRequestItem struct {
	ForUser uint `binding:"required" json:"for_user,omitempty"`
	Read    bool `binding:"required" json:"read,omitempty"`
	Write   bool `binding:"required" json:"write,omitempty"`
	Limited bool `binding:"required" json:"limited,omitempty"`
}

type MassDeleteRequest struct {
	Stocks bool `json:"stocks,omitempty"`
}

type StockMergeRequest struct {
	MergeInto uint `binding:"required" json:"merge_into,omitempty"`
	Stock     uint `binding:"required" json:"stock,omitempty"`
}

type CreateAccountRequest struct {
	Name       string `binding:"required" json:"name,omitempty"`
	UserID     uint   `binding:"required" json:"user_id,omitempty"`
	ProviderID uint   `binding:"required" json:"provider_id,omitempty"`
}
