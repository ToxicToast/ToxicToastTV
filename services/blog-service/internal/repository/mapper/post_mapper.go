package mapper

import (
	"toxictoast/services/blog-service/internal/domain"
	"toxictoast/services/blog-service/internal/repository/entity"

	"gorm.io/gorm"
)

// PostToEntity converts domain model to database entity
func PostToEntity(post *domain.Post) *entity.PostEntity {
	if post == nil {
		return nil
	}

	e := &entity.PostEntity{
		ID:              post.ID,
		Title:           post.Title,
		Slug:            post.Slug,
		Content:         post.Content,
		Excerpt:         post.Excerpt,
		Markdown:        post.Markdown,
		HTML:            post.HTML,
		Status:          string(post.Status),
		Featured:        post.Featured,
		AuthorID:        post.AuthorID,
		FeaturedImageID: post.FeaturedImageID,
		ReadingTime:     post.ReadingTime,
		ViewCount:       post.ViewCount,
		PublishedAt:     post.PublishedAt,
		CreatedAt:       post.CreatedAt,
		UpdatedAt:       post.UpdatedAt,
		MetaTitle:       post.MetaTitle,
		MetaDescription: post.MetaDescription,
		MetaKeywords:    post.MetaKeywords,
		OGTitle:         post.OGTitle,
		OGDescription:   post.OGDescription,
		OGImage:         post.OGImage,
		CanonicalURL:    post.CanonicalURL,
	}

	// Convert DeletedAt
	if post.DeletedAt != nil {
		e.DeletedAt = gorm.DeletedAt{
			Time:  *post.DeletedAt,
			Valid: true,
		}
	}

	// Convert Categories
	if len(post.Categories) > 0 {
		e.Categories = make([]entity.CategoryEntity, 0, len(post.Categories))
		for _, cat := range post.Categories {
			e.Categories = append(e.Categories, *CategoryToEntity(&cat))
		}
	}

	// Convert Tags
	if len(post.Tags) > 0 {
		e.Tags = make([]entity.TagEntity, 0, len(post.Tags))
		for _, tag := range post.Tags {
			e.Tags = append(e.Tags, *TagToEntity(&tag))
		}
	}

	return e
}

// PostToDomain converts database entity to domain model
func PostToDomain(e *entity.PostEntity) *domain.Post {
	if e == nil {
		return nil
	}

	post := &domain.Post{
		ID:              e.ID,
		Title:           e.Title,
		Slug:            e.Slug,
		Content:         e.Content,
		Excerpt:         e.Excerpt,
		Markdown:        e.Markdown,
		HTML:            e.HTML,
		Status:          domain.PostStatus(e.Status),
		Featured:        e.Featured,
		AuthorID:        e.AuthorID,
		FeaturedImageID: e.FeaturedImageID,
		ReadingTime:     e.ReadingTime,
		ViewCount:       e.ViewCount,
		PublishedAt:     e.PublishedAt,
		CreatedAt:       e.CreatedAt,
		UpdatedAt:       e.UpdatedAt,
		MetaTitle:       e.MetaTitle,
		MetaDescription: e.MetaDescription,
		MetaKeywords:    e.MetaKeywords,
		OGTitle:         e.OGTitle,
		OGDescription:   e.OGDescription,
		OGImage:         e.OGImage,
		CanonicalURL:    e.CanonicalURL,
	}

	// Convert DeletedAt
	if e.DeletedAt.Valid {
		deletedAt := e.DeletedAt.Time
		post.DeletedAt = &deletedAt
	}

	// Convert Categories
	if len(e.Categories) > 0 {
		post.Categories = make([]domain.Category, 0, len(e.Categories))
		for _, cat := range e.Categories {
			post.Categories = append(post.Categories, *CategoryToDomain(&cat))
		}
	}

	// Convert Tags
	if len(e.Tags) > 0 {
		post.Tags = make([]domain.Tag, 0, len(e.Tags))
		for _, tag := range e.Tags {
			post.Tags = append(post.Tags, *TagToDomain(&tag))
		}
	}

	return post
}

// PostsToDomain converts slice of entities to domain models
func PostsToDomain(entities []entity.PostEntity) []*domain.Post {
	posts := make([]*domain.Post, 0, len(entities))
	for _, e := range entities {
		posts = append(posts, PostToDomain(&e))
	}
	return posts
}
