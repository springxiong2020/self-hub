package mine

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"

	"github.com/springbear2020/self-hub/server/model/mine/dto"
)

type MineResourcesService struct{}

func (svc *MineResourcesService) SearchMineResources(uid uint, keyword string) ([]*dto.MineResourcesDTO, error) {
	if keyword == "" {
		return nil, fmt.Errorf("搜索关键词不能为空")
	}

	// 每个搜索任务对应一个固定的结果位置
	lists := make([][]*dto.MineResourcesDTO, 5)

	g, _ := errgroup.WithContext(context.Background())

	// 1. Search blogs
	g.Go(func() error {
		list, err := mineBlogsService.SearchMineBlogs(uid, keyword)
		if err != nil {
			return fmt.Errorf("blogs search error: %w", err)
		}
		lists[0] = list
		return nil
	})

	// 2. Search projects
	g.Go(func() error {
		list, err := mineProjectsService.SearchMineProjects(uid, keyword)
		if err != nil {
			return fmt.Errorf("projects search error: %w", err)
		}
		lists[1] = list
		return nil
	})

	// 3. Search books
	g.Go(func() error {
		list, err := mineBooksService.SearchMineBooks(uid, keyword)
		if err != nil {
			return fmt.Errorf("books search error: %w", err)
		}
		lists[2] = list
		return nil
	})

	// 4. Search sentences
	g.Go(func() error {
		list, err := mineSentencesService.SearchMineSentences(uid, keyword)
		if err != nil {
			return fmt.Errorf("sentences search error: %w", err)
		}
		lists[3] = list
		return nil
	})

	// 5. Search archives
	g.Go(func() error {
		list, err := mineArchivesService.SearchMineArchives(uid, keyword)
		if err != nil {
			return fmt.Errorf("archives search error: %w", err)
		}
		lists[4] = list
		return nil
	})

	// 等待所有搜索任务完成
	if err := g.Wait(); err != nil {
		return nil, err
	}

	// 按照固定索引顺序合并结果
	var results []*dto.MineResourcesDTO
	for _, list := range lists {
		results = append(results, list...)
	}

	return results, nil
}
