package scanc

import (
	"fmt"
	"time"
	"tprgdb/components/lrucache"
	"tprgdb/engine/constants"
	"tprgdb/engine/core"
	itemPackage "tprgdb/engine/item"
)

type RangeScanCursor struct {
	core         *core.Core
	keyRange     *Range
	pageSize     int
	currentPage  int
	currentItems []*itemPackage.DBItem
	cache        *lrucache.LRUCache[*Range]
}

func NewRangeScanCursor(core *core.Core, originalLowerBound, originalUpperBound string, pageSize int) *RangeScanCursor {
	return &RangeScanCursor{
		keyRange: &Range{
			LowerBound: constants.DATA_PREFIX + originalLowerBound,
			UpperBound: constants.DATA_PREFIX + originalUpperBound,
		},
		currentPage:  0,
		pageSize:     pageSize,
		currentItems: make([]*itemPackage.DBItem, 0, pageSize),
		cache:        lrucache.NewLRUCache[*Range](10), // Example cache size
		core:         core,
	}
}

func (c *RangeScanCursor) GetPage(pageNumber int) ([]*itemPackage.DBItem, error) {
	if pageNumber <= 0 {
		return nil, fmt.Errorf("page number invalid: %d", pageNumber)
	}

	if pageNumber == c.currentPage {
		return c.currentItems, nil
	}

	c.currentPage = pageNumber
	c.currentItems = make([]*itemPackage.DBItem, 0, c.pageSize)

	// If we have cached the range for this page number, use it
	if r, exists := c.cache.GetElement(fmt.Sprintf("%d", pageNumber)); exists {
		// We have the upper and lower bounds cached
		candidateKeys, err := c.core.GetKeysRange(r.LowerBound, r.UpperBound, 1, c.pageSize)
		if err != nil {
			return nil, fmt.Errorf("error getting keys range: %w", err)
		}
		items, err := c.core.GetMany(candidateKeys)
		if err != nil {
			return nil, fmt.Errorf("error getting items: %w", err)
		}
		for _, item := range items {
			if item != nil && !item.IsDeleted() {
				timestamp := time.Unix(int64(item.GetTimestamp()), 0)
				c.currentItems = append(c.currentItems, &itemPackage.DBItem{
					Key:       item.GetKey()[constants.PREFIX_SIZE:], // Remove the prefix
					Value:     item.GetValue(),
					Timestamp: &timestamp,
				})
			}
		}
		return c.currentItems, nil
	}

	// We have to calculate the range for this page number

	lowerBound := c.keyRange.LowerBound
	upperBound := c.keyRange.UpperBound

	startIndex := (pageNumber - 1) * c.pageSize
	count := 0
	pageCount := 1

	firstIteration := true

	for len(c.currentItems) < c.pageSize {

		candidateKeys, err := c.core.GetKeysRange(lowerBound, upperBound, 1, c.pageSize)
		if err != nil {
			return nil, fmt.Errorf("error getting keys range: %w", err)
		}
		if len(candidateKeys) == 0 {
			break // No more keys in the range
		}
		// Skip the first key if not the first iteration since it was already processed in the previous iteration
		if !firstIteration {
			candidateKeys = candidateKeys[1:]
		}
		if len(candidateKeys) == 0 {
			break // No more keys to process
		}
		firstIteration = false
		lowerBound = candidateKeys[len(candidateKeys)-1] // In next iteration start from the last key
		items, err := c.core.GetMany(candidateKeys)
		if err != nil {
			return nil, fmt.Errorf("error getting items: %w", err)
		}
		for _, item := range items {
			if item != nil && !item.IsDeleted() {
				count++
				if count-1 < startIndex {
					continue // Skip items until we reach the start index
				}
				if len(c.currentItems) >= c.pageSize {
					break // We have enough items for this page
				}
				timestamp := time.Unix(int64(item.GetTimestamp()), 0)
				c.currentItems = append(c.currentItems, &itemPackage.DBItem{
					Key:       item.GetKey()[constants.PREFIX_SIZE:], // Remove the prefix
					Value:     item.GetValue(),
					Timestamp: &timestamp,
				})
				if len(c.currentItems) == c.pageSize {
					break
				}
			}
			if len(c.currentItems) >= c.pageSize {
				break // We have enough items for this page
			}
		}
		pageCount++
		// Cache the range for this page number
	}
	if len(c.currentItems) > 0 {
		c.cache.AddElement(fmt.Sprintf("%d", pageNumber), &Range{
			LowerBound: c.currentItems[0].GetKey(),
			UpperBound: c.currentItems[len(c.currentItems)-1].GetKey(),
		})
	}
	return c.currentItems, nil

}

func (c *RangeScanCursor) GetCurrentPage() int {
	return c.currentPage
}

func (c *RangeScanCursor) GetPageSize() int {
	return c.pageSize
}

func (c *RangeScanCursor) NextPage() ([]*itemPackage.DBItem, error) {
	return c.GetPage(c.currentPage + 1)
}
func (c *RangeScanCursor) PreviousPage() ([]*itemPackage.DBItem, error) {
	if c.currentPage <= 1 {
		return nil, fmt.Errorf("no previous page")
	}
	return c.GetPage(c.currentPage - 1)
}
