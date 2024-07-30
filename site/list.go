package site

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gosimple/slug"
	"github.com/iand/genster/logging"
	"github.com/iand/genster/model"
	"github.com/iand/genster/render/md"
	"github.com/iand/genster/text"
)

func (s *Site) WriteAnomalyListPages(root string) error {
	baseDir := filepath.Join(root, s.ListAnomaliesDir)
	pn := NewPaginator()
	pn.HugoStyle = s.GenerateHugo
	for _, p := range s.Tree.People {
		if s.LinkFor(p) == "" {
			continue
		}
		if p.Redacted {
			logging.Debug("not writing redacted person to anomalies index", "id", p.ID)
			continue
		}
		categories := make([]model.AnomalyCategory, 0)
		anomaliesByCategory := make(map[model.AnomalyCategory][]*model.Anomaly)

		for _, a := range p.Anomalies {
			a := a // avoid shadowing
			al, ok := anomaliesByCategory[a.Category]
			if ok {
				al = append(al, a)
				anomaliesByCategory[a.Category] = al
				continue
			}

			categories = append(categories, a.Category)
			anomaliesByCategory[a.Category] = []*model.Anomaly{a}
		}
		sort.Slice(categories, func(i, j int) bool { return categories[i] < categories[j] })

		if len(anomaliesByCategory) > 0 {
			b := s.NewMarkdownBuilder()
			b.Heading2(md.Text(p.PreferredUniqueName), p.ID)
			rel := "unknown relation"
			if p.RelationToKeyPerson != nil && !p.RelationToKeyPerson.IsSelf() {
				rel = p.RelationToKeyPerson.Name()
			}

			links := b.EncodeModelLink("View page", p).String()

			// if p.EditLink != nil {
			// 	links += " or " + string(b.EncodeLink(text.LowerFirst(p.EditLink.Title), p.EditLink.URL))
			// }
			b.Para(b.EncodeText(text.FormatSentence(rel) + " " + links))
			for _, cat := range categories {
				al := anomaliesByCategory[cat]
				items := make([][2]md.Text, 0, len(al))

				for _, a := range al {
					items = append(items, [2]md.Text{
						md.Text(a.Context),
						md.Text(a.Text),
					})
				}
				b.DefinitionList(items)
			}

			group, groupPriority := groupRelation(p.RelationToKeyPerson)
			pn.AddEntryWithGroup(p.PreferredSortName+"~"+p.ID, p.PreferredSortName, b.String(), group, groupPriority)
		}

	}

	if err := pn.WritePages(s, baseDir, PageLayoutListAnomalies, "Anomalies", "Anomalies are errors or inconsistencies that have been detected in the underlying data used to generate this site."); err != nil {
		return err
	}

	return nil
}

func (s *Site) WriteInferenceListPages(root string) error {
	baseDir := filepath.Join(root, s.ListInferencesDir)
	pn := NewPaginator()
	pn.HugoStyle = s.GenerateHugo
	for _, p := range s.Tree.People {
		if s.LinkFor(p) == "" {
			continue
		}
		if p.Redacted {
			logging.Debug("not writing redacted person to inference index", "id", p.ID)
			continue
		}
		items := make([][2]md.Text, 0)
		for _, inf := range p.Inferences {
			items = append(items, [2]md.Text{
				md.Text(inf.Type + " " + inf.Value),
				md.Text("because " + inf.Reason),
			})
		}

		if len(items) > 0 {
			b := s.NewMarkdownBuilder()
			b.Heading2(md.Text(p.PreferredUniqueName), p.ID)
			// if p.EditLink != nil {
			// 	b.Para(render.Markdown(b.EncodeModelLink("View page", p) + " or " + string(b.EncodeLink(text.LowerFirst(p.EditLink.Title), p.EditLink.URL))))
			// } else {
			b.Para(md.Text(b.EncodeModelLink("View page", p)))
			// }
			b.DefinitionList(items)
			pn.AddEntry(p.PreferredSortName+"~"+p.ID, p.PreferredSortName, b.String())
		}

	}
	if err := pn.WritePages(s, baseDir, PageLayoutListInferences, "Inferences Made", "Inferences refer to hints and suggestions that help fill in missing information in the family tree."); err != nil {
		return err
	}

	return nil
}

func (s *Site) WriteTodoListPages(root string) error {
	baseDir := filepath.Join(root, s.ListTodoDir)
	pn := NewPaginator()
	pn.HugoStyle = s.GenerateHugo
	for _, p := range s.Tree.People {
		if s.LinkFor(p) == "" {
			continue
		}
		if p.Redacted {
			logging.Debug("not writing redacted person to todo index", "id", p.ID)
			continue
		}
		categories := make([]model.ToDoCategory, 0)
		todosByCategory := make(map[model.ToDoCategory][]*model.ToDo)

		for _, a := range p.ToDos {
			a := a // avoid shadowing
			al, ok := todosByCategory[a.Category]
			if ok {
				al = append(al, a)
				todosByCategory[a.Category] = al
				continue
			}

			categories = append(categories, a.Category)
			todosByCategory[a.Category] = []*model.ToDo{a}
		}
		sort.Slice(categories, func(i, j int) bool {
			return categories[i] < categories[j]
		})

		if len(todosByCategory) > 0 {
			b := s.NewMarkdownBuilder()
			b.Heading2(md.Text(p.PreferredUniqueName), p.ID)
			rel := "unknown relation"
			if p.RelationToKeyPerson != nil && !p.RelationToKeyPerson.IsSelf() {
				rel = p.RelationToKeyPerson.Name()
			}

			links := b.EncodeModelLink("View page", p).String()

			// if p.EditLink != nil {
			// 	links += " or " + string(b.EncodeLink(text.LowerFirst(p.EditLink.Title), p.EditLink.URL))
			// }
			b.Para(b.EncodeText(text.FormatSentence(rel) + " " + links))

			for _, cat := range categories {
				al := todosByCategory[cat]
				items := make([][2]md.Text, 0, len(al))

				for _, a := range al {
					line := text.StripTerminator(text.UpperFirst(a.Goal))
					if a.Reason != "" {
						line += " (" + text.LowerFirst(a.Reason) + ")"
					} else {
						line = text.FinishSentence(line)
					}
					items = append(items, [2]md.Text{
						md.Text(a.Context),
						md.Text(line),
					})
				}
				b.DefinitionList(items)

				// for _, a := range al {
				// 	line := b.EncodeItalic(a.Context) + ": " + text.StripTerminator(text.LowerFirst(a.Goal))
				// 	if a.Reason != "" {
				// 		line += " (" + text.LowerFirst(a.Reason) + ")"
				// 	} else {
				// 		line = text.FinishSentence(line)
				// 	}
				// 	items = append(items, line)
				// }
				// b.Heading4(cat.String())
				// b.UnorderedList(items)
			}

			group, groupPriority := groupRelation(p.RelationToKeyPerson)
			pn.AddEntryWithGroup(p.PreferredSortName+"~"+p.ID, p.PreferredSortName, b.String(), group, groupPriority)
		}

	}

	if err := pn.WritePages(s, baseDir, PageLayoutListTodo, "To Do", "These suggested tasks and projects are loose ends or incomplete areas in the tree that need further research."); err != nil {
		return err
	}

	return nil
}

func (s *Site) WritePersonListPages(root string) error {
	baseDir := filepath.Join(root, s.ListPeopleDir)
	pn := NewPaginator()
	pn.HugoStyle = s.GenerateHugo

	mentioned := make(map[string]*model.Person)

	// Include all people mentioned
	for _, p := range s.PublishSet.People {
		mentioned[p.ID] = p
		if !p.Father.IsUnknown() {
			mentioned[p.Father.ID] = p.Father
		}
		if !p.Mother.IsUnknown() {
			mentioned[p.Mother.ID] = p.Mother
		}
		for _, m := range p.Children {
			if !m.IsUnknown() {
				mentioned[m.ID] = m
			}
		}
		for _, m := range p.Spouses {
			if !m.IsUnknown() {
				mentioned[m.ID] = m
			}
		}
	}

	for _, p := range mentioned {
		if p.Redacted {
			logging.Debug("not writing redacted person to person index", "id", p.ID)
			continue
		}
		items := make([][2]md.Text, 0)
		b := &CitationSkippingEncoder[md.Text]{s.NewMarkdownBuilder()}

		summary := PersonSummary(p, b, b.EncodeText(p.PreferredFamiliarName), true, true, false)

		var rel string
		if s.LinkFor(p) != "" {
			if p.RelationToKeyPerson != nil && !p.RelationToKeyPerson.IsSelf() {
				rel = b.EncodeBold(b.EncodeText(text.FormatSentence(p.RelationToKeyPerson.Name()))).String()
			}
		}

		items = append(items, [2]md.Text{
			b.EncodeText(text.AppendClause(b.EncodeModelLink(b.EncodeText(p.PreferredSortName), p).String(), rel)),
			summary,
		})
		b.DefinitionList(items)
		pn.AddEntry(p.PreferredSortName+"~"+p.ID, p.PreferredSortName, b.String())

	}
	if err := pn.WritePages(s, baseDir, PageLayoutListPeople, "People", "This is a full, alphabetical list of people in the tree."); err != nil {
		return err
	}
	return nil
}

func (s *Site) WritePlaceListPages(root string) error {
	baseDir := filepath.Join(root, s.ListPlacesDir)
	pn := NewPaginator()
	pn.HugoStyle = s.GenerateHugo
	for _, p := range s.PublishSet.Places {
		if len(p.Timeline) == 0 {
			continue
		}
		items := make([][2]md.Text, 0)
		b := s.NewMarkdownBuilder()
		items = append(items, [2]md.Text{
			md.Text(b.EncodeModelLink(b.EncodeText(p.PreferredUniqueName), p)),
		})
		b.DefinitionList(items)
		pn.AddEntry(p.PreferredSortName+"~"+p.ID, p.PreferredSortName, b.String())

	}
	if err := pn.WritePages(s, baseDir, PageLayoutListPlaces, "Places", "This is a full, alphabetical list of places in the tree."); err != nil {
		return err
	}
	return nil
}

func (s *Site) WriteSourceListPages(root string) error {
	baseDir := filepath.Join(root, s.ListSourcesDir)
	pn := NewPaginator()
	pn.HugoStyle = s.GenerateHugo
	for _, so := range s.PublishSet.Sources {
		items := make([][2]md.Text, 0)
		b := s.NewMarkdownBuilder()
		items = append(items, [2]md.Text{
			md.Text(b.EncodeModelLink(b.EncodeText(so.Title), so)),
		})
		b.DefinitionList(items)
		pn.AddEntry(so.Title+"~"+so.ID, so.Title, b.String())

	}
	if err := pn.WritePages(s, baseDir, PageLayoutListSources, "Sources", "This is a full, alphabetical list of sources cited in the tree."); err != nil {
		return err
	}

	return nil
}

func (s *Site) WriteSurnameListPages(root string) error {
	peopleBySurname := make(map[string][]*model.Person)
	for _, p := range s.PublishSet.People {
		if s.LinkFor(p) == "" || p.IsUnknown() {
			continue
		}
		if p.Redacted {
			logging.Debug("not writing redacted person to surname index", "id", p.ID)
			continue
		}
		if p.PreferredFamilyName == model.UnknownNamePlaceholder {
			continue
		}
		if !p.IsDirectAncestor() {
			continue
		}
		peopleBySurname[p.FamilyNameGrouping] = append(peopleBySurname[p.FamilyNameGrouping], p)
	}

	surnames := make([]string, 0, len(peopleBySurname))

	for surname, people := range peopleBySurname {
		surnames = append(surnames, surname)
		model.SortPeopleByGeneration(people)

		pn := NewPaginator()
		pn.HugoStyle = s.GenerateHugo
		pn.MaxPageSize = -1

		for _, p := range people {
			items := make([][2]md.Text, 0)
			b := &CitationSkippingEncoder[md.Text]{s.NewMarkdownBuilder()}

			title := b.EncodeModelLink(b.EncodeText(p.PreferredSortName), p)
			summary := PersonSummary(p, b, b.EncodeText(p.PreferredFamiliarName), true, true, false)

			var rel string
			if s.LinkFor(p) != "" {
				if p.RelationToKeyPerson != nil && !p.RelationToKeyPerson.IsSelf() {
					rel = b.EncodeBold(b.EncodeText(p.RelationToKeyPerson.Name())).String()
					title = b.EncodeText(text.AppendClause(title.String(), rel))
				}
			}

			if p.Olb != "" {
				summary = md.Text(b.EncodeItalic(b.EncodeText(text.FormatSentence(p.Olb)))+"<br>") + summary
			}

			items = append(items, [2]md.Text{
				b.EncodeText(text.FormatSentence(title.String())),
				summary,
			})

			b.DefinitionList(items)
			pn.AddEntry(fmt.Sprintf("%4d~%s~%s", p.Generation(), p.PreferredSortName, p.ID), p.PreferredSortName, b.String())
		}

		baseDir := filepath.Join(root, s.ListSurnamesDir, slug.Make(surname))

		desc := "This is a full, alphabetical list of ancestors with the surname " + surname + "."
		if strings.Contains(surname, "(") {
			clean := strings.Replace(surname, "(", " ", -1)
			clean = strings.Replace(clean, ")", " ", -1)
			clean = strings.Replace(clean, ",", " ", -1)
			surnames := strings.Fields(clean)
			desc = "This is a full, alphabetical list of ancestors with the surnames " + text.JoinListOr(surnames) + "."
		}

		if err := pn.WritePages(s, baseDir, PageLayoutListSurnames, surname, desc); err != nil {
			return err
		}

	}

	sort.Slice(surnames, func(i, j int) bool { return surnames[i] < surnames[j] })
	indexPage := "index.md"
	if s.GenerateHugo {
		indexPage = "_index.md"
	}

	doc := s.NewDocument()
	doc.Title("Surnames")
	doc.Summary("This is a full, alphabetical list of the surnames of ancestors in the tree.")
	doc.Layout(PageLayoutListSurnames.String())

	alist := make([]md.Text, 0, len(surnames))
	for _, surname := range surnames {
		alist = append(alist, doc.EncodeLink(doc.EncodeText(surname), s.LinkForSurnameListPage(surname)))
	}
	doc.UnorderedList(alist)

	baseDir := filepath.Join(root, s.ListSurnamesDir)
	if err := writePage(doc, baseDir, indexPage); err != nil {
		return fmt.Errorf("failed to write surname index: %w", err)
	}

	return nil
}
