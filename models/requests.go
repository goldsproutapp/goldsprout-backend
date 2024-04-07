package models


type StockSnapshotCreationPayload struct {

    ProviderID uint `binding:"required" json:"provider_id"`
    StockName string `binding:"required" json:"stock_name"`
    StockCode string `binding:"required" json:"stock_code"`

    Units string `binding:"required" json:"units"` 
    Price string `binding:"required" json:"price"` 
    Cost string `binding:"required" json:"cost"`
    Value string `binding:"required" json:"value"`

    AbsoluteChange string `binding:"required" json:"absolute_change"`
}

type StockSnapshotCreationRequest struct {
    Entries []StockSnapshotCreationPayload `binding:"required" json:"entries"`
    UserID uint `binding:"required" json:"user_id"`
    Date int64 `binding:"required" json:"date"`
    DeleteSoldStocks bool `json:"delete_sold_stocks"`
}

type StockUpdateRequest struct {
    Stock Stock `binding:"required" json:"stock"`
}

type ProviderUpdateRequest struct {
    Provider Provider `binding:"required" json:"provider"`
}

type PerformanceRequestQuery struct {
	Compare string `binding:"required" json:"compare,omitempty" form:"compare"`
	Of      string `binding:"required" json:"of,omitempty" form:"of"`
	For     string `binding:"required" json:"for,omitempty" form:"for"`
	Over    string `binding:"required" json:"over,omitempty" form:"over"`
}
