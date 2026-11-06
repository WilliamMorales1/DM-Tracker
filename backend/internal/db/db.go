package db

import (
	"context"
	"database/sql"
	_ "modernc.org/sqlite"
)

// Open opens the sqlite database at path with foreign keys enabled.
func Open(path string) (*sql.DB, error) {
	return sql.Open("sqlite", path+"?_pragma=foreign_keys(1)")
}

// OpenReadOnly opens the sqlite database at path with query_only enabled, so
// the connection rejects any write regardless of what SQL text it's handed.
// Meant for untrusted query sources (e.g. the chat agent's SQL tool).
func OpenReadOnly(path string) (*sql.DB, error) {
	return sql.Open("sqlite", path+"?_pragma=query_only(1)")
}

type DBTX interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func New(db DBTX) *Queries {
	return &Queries{db: db}
}

type Queries struct {
	db DBTX
}

func (q *Queries) WithTx(tx *sql.Tx) *Queries {
	return &Queries{
		db: tx,
	}
}

type Monster struct {
	ID                  int64
	Name                string
	Type                sql.NullString
	Cr                  sql.NullString
	Hp                  sql.NullInt64
	HpFormula           sql.NullString
	Ac                  sql.NullInt64
	AcDesc              sql.NullString
	Speed               sql.NullString
	Str                 sql.NullInt64
	Dex                 sql.NullInt64
	Con                 sql.NullInt64
	IntScore            sql.NullInt64
	Wis                 sql.NullInt64
	Cha                 sql.NullInt64
	SavingThrows        sql.NullString
	Skills              sql.NullString
	DamageResistances   sql.NullString
	DamageImmunities    sql.NullString
	ConditionImmunities sql.NullString
	Senses              sql.NullString
	Languages           sql.NullString
	Traits              sql.NullString
	Actions             sql.NullString
	LegendaryActions    sql.NullString
	Notes               sql.NullString
	Source              sql.NullString
	Size                sql.NullString
	Alignment           sql.NullString
	PassivePerception   sql.NullInt64
	Vulnerabilities     sql.NullString
	Environment         sql.NullString
	ImageUrl            sql.NullString
	Reactions           sql.NullString
	Spellcasting        sql.NullString
	TokenUrl            sql.NullString
}

type Note struct {
	Name    string
	Content string
}

type InitiativePreset struct {
	Name       string
	Combatants string
}

type Session struct {
	ID         int64
	SessionNum int64
	Title      string
	Chapters   sql.NullString
	LevelStart sql.NullInt64
	LevelEnd   sql.NullInt64
	Summary    sql.NullString
	Checkpoint sql.NullString
}

type DemonLord struct {
	ID                int64
	Name              string
	Dominions         string
	Epithets          string
	Layer             string
	Description       string
	Servants          string
	Component         string
	ComponentLocation string
}

type Action struct {
	ID          int64
	Name        string
	Tag         string
	Description string
}

type SkillArea struct {
	ID    int64
	Skill string
	Areas string
}

type Condition struct {
	ID               int64
	Name             string
	Description      sql.NullString
	Effects          sql.NullString
	DescriptionAfter sql.NullString
}

type ExhaustionLevel struct {
	ID     int64
	Level  string
	Effect string
}

const getMonster = `
SELECT id, name, type, cr, hp, hp_formula, ac, ac_desc, speed,
  str, dex, con, int_score, wis, cha,
  COALESCE(saving_throws, '')        AS saving_throws,
  COALESCE(skills, '')               AS skills,
  COALESCE(damage_resistances, '')   AS damage_resistances,
  COALESCE(damage_immunities, '')    AS damage_immunities,
  COALESCE(vulnerabilities, '')      AS vulnerabilities,
  COALESCE(condition_immunities, '') AS condition_immunities,
  COALESCE(senses, '')               AS senses,
  passive_perception,
  COALESCE(languages, '')            AS languages,
  COALESCE(traits, '')               AS traits,
  COALESCE(actions, '')              AS actions,
  COALESCE(reactions, '')            AS reactions,
  COALESCE(legendary_actions, '')    AS legendary_actions,
  COALESCE(bonus_actions, '')        AS bonus_actions,
  COALESCE(spellcasting, '')         AS spellcasting,
  COALESCE(lair_actions, '')         AS lair_actions,
  COALESCE(regional_effects, '')     AS regional_effects,
  COALESCE(notes, '')                AS notes,
  COALESCE(source, '')               AS source,
  COALESCE(size, '')                 AS size,
  COALESCE(alignment, '')            AS alignment,
  COALESCE(environment, '')          AS environment,
  COALESCE(image_url, '')            AS image_url,
  COALESCE(token_url, '')            AS token_url
FROM Monsters WHERE id = ?
`

type GetMonsterRow struct {
	ID                  int64
	Name                string
	Type                sql.NullString
	Cr                  sql.NullString
	Hp                  sql.NullInt64
	HpFormula           sql.NullString
	Ac                  sql.NullInt64
	AcDesc              sql.NullString
	Speed               sql.NullString
	Str                 sql.NullInt64
	Dex                 sql.NullInt64
	Con                 sql.NullInt64
	IntScore            sql.NullInt64
	Wis                 sql.NullInt64
	Cha                 sql.NullInt64
	SavingThrows        string
	Skills              string
	DamageResistances   string
	DamageImmunities    string
	Vulnerabilities     string
	ConditionImmunities string
	Senses              string
	PassivePerception   sql.NullInt64
	Languages           string
	Traits              string
	Actions             string
	Reactions           string
	LegendaryActions    string
	BonusActions        string
	Spellcasting        string
	LairActions         string
	RegionalEffects     string
	Notes               string
	Source              string
	Size                string
	Alignment           string
	Environment         string
	ImageUrl            string
	TokenUrl            string
}

func (q *Queries) GetMonster(ctx context.Context, id int64) (GetMonsterRow, error) {
	row := q.db.QueryRowContext(ctx, getMonster, id)
	var i GetMonsterRow
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Type,
		&i.Cr,
		&i.Hp,
		&i.HpFormula,
		&i.Ac,
		&i.AcDesc,
		&i.Speed,
		&i.Str,
		&i.Dex,
		&i.Con,
		&i.IntScore,
		&i.Wis,
		&i.Cha,
		&i.SavingThrows,
		&i.Skills,
		&i.DamageResistances,
		&i.DamageImmunities,
		&i.Vulnerabilities,
		&i.ConditionImmunities,
		&i.Senses,
		&i.PassivePerception,
		&i.Languages,
		&i.Traits,
		&i.Actions,
		&i.Reactions,
		&i.LegendaryActions,
		&i.BonusActions,
		&i.Spellcasting,
		&i.LairActions,
		&i.RegionalEffects,
		&i.Notes,
		&i.Source,
		&i.Size,
		&i.Alignment,
		&i.Environment,
		&i.ImageUrl,
		&i.TokenUrl,
	)
	return i, err
}

const getNote = `SELECT name, content FROM Notes WHERE name = ?`

func (q *Queries) GetNote(ctx context.Context, name string) (Note, error) {
	row := q.db.QueryRowContext(ctx, getNote, name)
	var i Note
	err := row.Scan(&i.Name, &i.Content)
	return i, err
}

const listMonsterStats = `SELECT id, name, ac, hp, dex FROM Monsters ORDER BY name`

type ListMonsterStatsRow struct {
	ID   int64
	Name string
	Ac   sql.NullInt64
	Hp   sql.NullInt64
	Dex  sql.NullInt64
}

func (q *Queries) ListMonsterStats(ctx context.Context) ([]ListMonsterStatsRow, error) {
	rows, err := q.db.QueryContext(ctx, listMonsterStats)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListMonsterStatsRow
	for rows.Next() {
		var i ListMonsterStatsRow
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Ac,
			&i.Hp,
			&i.Dex,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listMonsters = `
SELECT id, name, type, cr, COALESCE(image_url, '') AS image_url, COALESCE(source, '') AS source
FROM Monsters
ORDER BY
  CASE cr WHEN '1/8' THEN 0.125 WHEN '1/4' THEN 0.25 WHEN '1/2' THEN 0.5
    ELSE CAST(cr AS REAL) END,
  name
`

type ListMonstersRow struct {
	ID       int64
	Name     string
	Type     sql.NullString
	Cr       sql.NullString
	ImageUrl string
	Source   string
}

func (q *Queries) ListMonsters(ctx context.Context) ([]ListMonstersRow, error) {
	rows, err := q.db.QueryContext(ctx, listMonsters)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListMonstersRow
	for rows.Next() {
		var i ListMonstersRow
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Type,
			&i.Cr,
			&i.ImageUrl,
			&i.Source,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listNoteNames = `SELECT name FROM Notes ORDER BY name`

func (q *Queries) ListNoteNames(ctx context.Context) ([]string, error) {
	rows, err := q.db.QueryContext(ctx, listNoteNames)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		items = append(items, name)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listNotes = `SELECT name, content FROM Notes ORDER BY name`

func (q *Queries) ListNotes(ctx context.Context) ([]Note, error) {
	rows, err := q.db.QueryContext(ctx, listNotes)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Note
	for rows.Next() {
		var i Note
		if err := rows.Scan(&i.Name, &i.Content); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listInitiativePresets = `SELECT name, combatants FROM InitiativePresets ORDER BY name`

func (q *Queries) ListInitiativePresets(ctx context.Context) ([]InitiativePreset, error) {
	rows, err := q.db.QueryContext(ctx, listInitiativePresets)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []InitiativePreset
	for rows.Next() {
		var i InitiativePreset
		if err := rows.Scan(&i.Name, &i.Combatants); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const deleteInitiativePreset = `DELETE FROM InitiativePresets WHERE name = ?`

func (q *Queries) DeleteInitiativePreset(ctx context.Context, name string) error {
	_, err := q.db.ExecContext(ctx, deleteInitiativePreset, name)
	return err
}

const listDemonLords = `
SELECT id, name, dominions, epithets, layer, description, servants, component, component_location
FROM DemonLords ORDER BY name
`

func (q *Queries) ListDemonLords(ctx context.Context) ([]DemonLord, error) {
	rows, err := q.db.QueryContext(ctx, listDemonLords)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []DemonLord
	for rows.Next() {
		var i DemonLord
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Dominions,
			&i.Epithets,
			&i.Layer,
			&i.Description,
			&i.Servants,
			&i.Component,
			&i.ComponentLocation,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

const listActions = `
SELECT id, name, tag, description FROM Actions ORDER BY name
`

func (q *Queries) ListActions(ctx context.Context) ([]Action, error) {
	rows, err := q.db.QueryContext(ctx, listActions)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Action
	for rows.Next() {
		var i Action
		if err := rows.Scan(&i.ID, &i.Name, &i.Tag, &i.Description); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

const listSkillAreas = `
SELECT id, skill, areas FROM SkillAreas ORDER BY skill
`

func (q *Queries) ListSkillAreas(ctx context.Context) ([]SkillArea, error) {
	rows, err := q.db.QueryContext(ctx, listSkillAreas)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SkillArea
	for rows.Next() {
		var i SkillArea
		if err := rows.Scan(&i.ID, &i.Skill, &i.Areas); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

const listConditions = `
SELECT id, name, description, effects, description_after FROM Conditions ORDER BY id
`

func (q *Queries) ListConditions(ctx context.Context) ([]Condition, error) {
	rows, err := q.db.QueryContext(ctx, listConditions)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Condition
	for rows.Next() {
		var i Condition
		if err := rows.Scan(&i.ID, &i.Name, &i.Description, &i.Effects, &i.DescriptionAfter); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

const listExhaustionLevels = `
SELECT id, level, effect FROM ExhaustionLevels ORDER BY id
`

func (q *Queries) ListExhaustionLevels(ctx context.Context) ([]ExhaustionLevel, error) {
	rows, err := q.db.QueryContext(ctx, listExhaustionLevels)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ExhaustionLevel
	for rows.Next() {
		var i ExhaustionLevel
		if err := rows.Scan(&i.ID, &i.Level, &i.Effect); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

const listSessions = `
SELECT id, session_num, title, chapters, level_start, level_end, summary, checkpoint
FROM Sessions ORDER BY session_num
`

func (q *Queries) ListSessions(ctx context.Context) ([]Session, error) {
	rows, err := q.db.QueryContext(ctx, listSessions)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Session
	for rows.Next() {
		var i Session
		if err := rows.Scan(
			&i.ID,
			&i.SessionNum,
			&i.Title,
			&i.Chapters,
			&i.LevelStart,
			&i.LevelEnd,
			&i.Summary,
			&i.Checkpoint,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const upsertMonster = `
INSERT INTO Monsters (
  name, type, cr, hp, hp_formula, ac, ac_desc, speed,
  str, dex, con, int_score, wis, cha,
  saving_throws, skills, damage_resistances, damage_immunities, vulnerabilities,
  condition_immunities, senses, passive_perception, languages,
  traits, actions, reactions, legendary_actions, bonus_actions, spellcasting,
  lair_actions, regional_effects,
  source, size, alignment, environment, image_url, token_url
) VALUES (
  ?, ?, ?, ?, ?, ?, ?, ?,
  ?, ?, ?, ?, ?, ?,
  ?, ?, ?, ?, ?,
  ?, ?, ?, ?,
  ?, ?, ?, ?, ?, ?,
  ?, ?,
  ?, ?, ?, ?, ?, ?
)
ON CONFLICT(name) DO UPDATE SET
  type = excluded.type, cr = excluded.cr, hp = excluded.hp, hp_formula = excluded.hp_formula,
  ac = excluded.ac, ac_desc = excluded.ac_desc, speed = excluded.speed,
  str = excluded.str, dex = excluded.dex, con = excluded.con, int_score = excluded.int_score,
  wis = excluded.wis, cha = excluded.cha,
  saving_throws = excluded.saving_throws, skills = excluded.skills,
  damage_resistances = excluded.damage_resistances, damage_immunities = excluded.damage_immunities,
  vulnerabilities = excluded.vulnerabilities, condition_immunities = excluded.condition_immunities,
  senses = excluded.senses, passive_perception = excluded.passive_perception, languages = excluded.languages,
  traits = excluded.traits, actions = excluded.actions, reactions = excluded.reactions,
  legendary_actions = excluded.legendary_actions, bonus_actions = excluded.bonus_actions, spellcasting = excluded.spellcasting,
  lair_actions = excluded.lair_actions, regional_effects = excluded.regional_effects,
  source = excluded.source, size = excluded.size, alignment = excluded.alignment,
  environment = excluded.environment, image_url = excluded.image_url, token_url = excluded.token_url
`

type UpsertMonsterParams struct {
	Name                string
	Type                sql.NullString
	Cr                  sql.NullString
	Hp                  sql.NullInt64
	HpFormula           sql.NullString
	Ac                  sql.NullInt64
	AcDesc              sql.NullString
	Speed               sql.NullString
	Str                 sql.NullInt64
	Dex                 sql.NullInt64
	Con                 sql.NullInt64
	IntScore            sql.NullInt64
	Wis                 sql.NullInt64
	Cha                 sql.NullInt64
	SavingThrows        sql.NullString
	Skills              sql.NullString
	DamageResistances   sql.NullString
	DamageImmunities    sql.NullString
	Vulnerabilities     sql.NullString
	ConditionImmunities sql.NullString
	Senses              sql.NullString
	PassivePerception   sql.NullInt64
	Languages           sql.NullString
	Traits              sql.NullString
	Actions             sql.NullString
	Reactions           sql.NullString
	LegendaryActions    sql.NullString
	BonusActions        sql.NullString
	Spellcasting        sql.NullString
	LairActions         sql.NullString
	RegionalEffects     sql.NullString
	Source              sql.NullString
	Size                sql.NullString
	Alignment           sql.NullString
	Environment         sql.NullString
	ImageUrl            sql.NullString
	TokenUrl            sql.NullString
}

func (q *Queries) UpsertMonster(ctx context.Context, arg UpsertMonsterParams) error {
	_, err := q.db.ExecContext(ctx, upsertMonster,
		arg.Name,
		arg.Type,
		arg.Cr,
		arg.Hp,
		arg.HpFormula,
		arg.Ac,
		arg.AcDesc,
		arg.Speed,
		arg.Str,
		arg.Dex,
		arg.Con,
		arg.IntScore,
		arg.Wis,
		arg.Cha,
		arg.SavingThrows,
		arg.Skills,
		arg.DamageResistances,
		arg.DamageImmunities,
		arg.Vulnerabilities,
		arg.ConditionImmunities,
		arg.Senses,
		arg.PassivePerception,
		arg.Languages,
		arg.Traits,
		arg.Actions,
		arg.Reactions,
		arg.LegendaryActions,
		arg.BonusActions,
		arg.Spellcasting,
		arg.LairActions,
		arg.RegionalEffects,
		arg.Source,
		arg.Size,
		arg.Alignment,
		arg.Environment,
		arg.ImageUrl,
		arg.TokenUrl,
	)
	return err
}

const upsertNote = `
INSERT INTO Notes (name, content) VALUES (?, ?)
ON CONFLICT(name) DO UPDATE SET content = excluded.content
`

type UpsertNoteParams struct {
	Name    string
	Content string
}

func (q *Queries) UpsertNote(ctx context.Context, arg UpsertNoteParams) error {
	_, err := q.db.ExecContext(ctx, upsertNote, arg.Name, arg.Content)
	return err
}

const upsertInitiativePreset = `
INSERT INTO InitiativePresets (name, combatants) VALUES (?, ?)
ON CONFLICT(name) DO UPDATE SET combatants = excluded.combatants
`

type UpsertInitiativePresetParams struct {
	Name       string
	Combatants string
}

func (q *Queries) UpsertInitiativePreset(ctx context.Context, arg UpsertInitiativePresetParams) error {
	_, err := q.db.ExecContext(ctx, upsertInitiativePreset, arg.Name, arg.Combatants)
	return err
}

const upsertSpell = `
INSERT INTO Spells (
  name, level, school, ritual, casting_time, range, components, duration,
  concentration, classes, description, higher_level, source
) VALUES (
  ?, ?, ?, ?, ?, ?, ?, ?,
  ?, ?, ?, ?, ?
)
ON CONFLICT(name) DO UPDATE SET
  level = excluded.level, school = excluded.school, ritual = excluded.ritual,
  casting_time = excluded.casting_time, range = excluded.range,
  components = excluded.components, duration = excluded.duration,
  concentration = excluded.concentration, classes = excluded.classes,
  description = excluded.description, higher_level = excluded.higher_level,
  source = excluded.source
`

type UpsertSpellParams struct {
	Name          string
	Level         int64
	School        sql.NullString
	Ritual        bool
	CastingTime   sql.NullString
	Range         sql.NullString
	Components    sql.NullString
	Duration      sql.NullString
	Concentration bool
	Classes       sql.NullString
	Description   sql.NullString
	HigherLevel   sql.NullString
	Source        sql.NullString
}

func (q *Queries) UpsertSpell(ctx context.Context, arg UpsertSpellParams) error {
	_, err := q.db.ExecContext(ctx, upsertSpell,
		arg.Name,
		arg.Level,
		arg.School,
		arg.Ritual,
		arg.CastingTime,
		arg.Range,
		arg.Components,
		arg.Duration,
		arg.Concentration,
		arg.Classes,
		arg.Description,
		arg.HigherLevel,
		arg.Source,
	)
	return err
}

const listSpells = `SELECT id, name, level, school, COALESCE(classes, '') AS classes, COALESCE(source, '') AS source FROM Spells ORDER BY level, name`

type ListSpellsRow struct {
	ID      int64
	Name    string
	Level   int64
	School  sql.NullString
	Classes string
	Source  string
}

func (q *Queries) ListSpells(ctx context.Context) ([]ListSpellsRow, error) {
	rows, err := q.db.QueryContext(ctx, listSpells)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListSpellsRow
	for rows.Next() {
		var i ListSpellsRow
		if err := rows.Scan(&i.ID, &i.Name, &i.Level, &i.School, &i.Classes, &i.Source); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getSpell = `
SELECT id, name, level,
  COALESCE(school, '')        AS school,
  ritual,
  COALESCE(casting_time, '')  AS casting_time,
  COALESCE(range, '')         AS range,
  COALESCE(components, '')    AS components,
  COALESCE(duration, '')      AS duration,
  concentration,
  COALESCE(classes, '')       AS classes,
  COALESCE(description, '')   AS description,
  COALESCE(higher_level, '')  AS higher_level,
  COALESCE(source, '')        AS source
FROM Spells WHERE id = ?
`

type GetSpellRow struct {
	ID            int64
	Name          string
	Level         int64
	School        string
	Ritual        bool
	CastingTime   string
	Range         string
	Components    string
	Duration      string
	Concentration bool
	Classes       string
	Description   string
	HigherLevel   string
	Source        string
}

func (q *Queries) GetSpell(ctx context.Context, id int64) (GetSpellRow, error) {
	row := q.db.QueryRowContext(ctx, getSpell, id)
	var i GetSpellRow
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Level,
		&i.School,
		&i.Ritual,
		&i.CastingTime,
		&i.Range,
		&i.Components,
		&i.Duration,
		&i.Concentration,
		&i.Classes,
		&i.Description,
		&i.HigherLevel,
		&i.Source,
	)
	return i, err
}
