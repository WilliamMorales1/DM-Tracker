CREATE TABLE Monsters (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    type TEXT,
    cr TEXT,
    hp INT,
    hp_formula TEXT,
    ac INT,
    ac_desc TEXT,
    speed TEXT,
    str INT DEFAULT 10,
    dex INT DEFAULT 10,
    con INT DEFAULT 10,
    int_score INT DEFAULT 10,
    wis INT DEFAULT 10,
    cha INT DEFAULT 10,
    saving_throws TEXT,
    skills TEXT,
    damage_resistances TEXT,
    damage_immunities TEXT,
    condition_immunities TEXT,
    senses TEXT,
    languages TEXT,
    traits TEXT,
    actions TEXT,
    legendary_actions TEXT,
    notes TEXT,
    source TEXT,
    size TEXT,
    alignment TEXT,
    passive_perception INT,
    vulnerabilities TEXT,
    environment TEXT,
    image_url TEXT,
    reactions TEXT,
    spellcasting TEXT,
    token_url TEXT,
    bonus_actions TEXT,
    lair_actions TEXT,
    regional_effects TEXT
);

CREATE TABLE Notes (
    name TEXT PRIMARY KEY,
    content TEXT NOT NULL DEFAULT ''
);

CREATE TABLE GameMaps (
    id TEXT PRIMARY KEY,
    img TEXT NOT NULL
);

CREATE TABLE MapMarkers (
    map_id TEXT REFERENCES GameMaps(id) ON DELETE CASCADE,
    i INT NOT NULL,
    x INT NOT NULL,
    y INT NOT NULL,
    title TEXT NOT NULL,
    body TEXT NOT NULL CHECK (length(body) <= 500),
    PRIMARY KEY (map_id, i)
);

CREATE TABLE chapter_chunks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    chapter_title TEXT NOT NULL,
    content TEXT NOT NULL,
    embedding TEXT NOT NULL
);

CREATE VIRTUAL TABLE chapter_chunks_fts USING fts5(
    chapter_title,
    content,
    content='chapter_chunks',
    content_rowid='id',
    tokenize='porter unicode61'
);

CREATE TRIGGER chapter_chunks_ai AFTER INSERT ON chapter_chunks BEGIN
    INSERT INTO chapter_chunks_fts(rowid, chapter_title, content)
    VALUES (new.id, new.chapter_title, new.content);
END;

CREATE TRIGGER chapter_chunks_ad AFTER DELETE ON chapter_chunks BEGIN
    INSERT INTO chapter_chunks_fts(chapter_chunks_fts, rowid, chapter_title, content)
    VALUES ('delete', old.id, old.chapter_title, old.content);
END;

CREATE TRIGGER chapter_chunks_au AFTER UPDATE ON chapter_chunks BEGIN
    INSERT INTO chapter_chunks_fts(chapter_chunks_fts, rowid, chapter_title, content)
    VALUES ('delete', old.id, old.chapter_title, old.content);
    INSERT INTO chapter_chunks_fts(rowid, chapter_title, content)
    VALUES (new.id, new.chapter_title, new.content);
END;

CREATE TABLE Spells (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    level INT NOT NULL,
    school TEXT,
    ritual INT NOT NULL DEFAULT 0,
    casting_time TEXT,
    range TEXT,
    components TEXT,
    duration TEXT,
    concentration INT NOT NULL DEFAULT 0,
    classes TEXT,
    description TEXT,
    higher_level TEXT,
    source TEXT
);

CREATE TABLE InitiativePresets (
    name TEXT PRIMARY KEY,
    combatants TEXT NOT NULL DEFAULT '[]'
);

CREATE TABLE Actions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    tag TEXT NOT NULL,
    description TEXT NOT NULL
);

INSERT INTO Actions (name, tag, description) VALUES
('Study', 'ACTION', 'When you take the Study action, make an Intelligence check to recall useful information about a source of knowledge.');

CREATE TABLE SkillAreas (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    skill TEXT NOT NULL,
    areas TEXT NOT NULL
);

INSERT INTO SkillAreas (skill, areas) VALUES
('Arcana', 'Spells, magic items, magical traditions, planes, and certain creatures'),
('History', 'Historic events, people, civilizations, and wars'),
('Investigation', 'Traps, ciphers, riddles, and gadgetry'),
('Nature', 'Terrain, flora, weather, and certain creatures'),
('Religion', 'Deities, rites, holy symbols, cults, and certain creatures');

CREATE TABLE Conditions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    effects TEXT,
    description_after TEXT
);

INSERT INTO Conditions (name, effects) VALUES
('Blinded', 'A blinded creature can''t see and automatically fails any ability check that requires sight.
Attack rolls against the creature have advantage, and the creature''s attack rolls have disadvantage.'),
('Charmed', 'A charmed creature can''t attack the charmer or target the charmer with harmful abilities or magical effects.
The charmer has advantage on any ability check to interact socially with the creature.'),
('Deafened', 'A deafened creature can''t hear and automatically fails any ability check that requires hearing.'),
('Frightened', 'A frightened creature has disadvantage on ability checks and attack rolls while the source of its fear is within line of sight.
The creature can''t willingly move closer to the source of its fear.'),
('Grappled', 'A grappled creature''s speed becomes 0, and it can''t benefit from any bonus to its speed.
The condition ends if the grappler is incapacitated (see the condition).
The condition also ends if an effect removes the grappled creature from the reach of the grappler or grappling effect, such as when a creature is hurled away by the thunderwave spell.'),
('Incapacitated', 'An incapacitated creature can''t take actions or reactions.'),
('Invisible', 'An invisible creature is impossible to see without the aid of magic or a special sense. For the purpose of hiding, the creature is heavily obscured. The creature''s location can be detected by any noise it makes or any tracks it leaves.
Attack rolls against the creature have disadvantage, and the creature''s attack rolls have advantage.'),
('Paralyzed', 'A paralyzed creature is incapacitated (see the condition) and can''t move or speak.
The creature automatically fails Strength and Dexterity saving throws.
Attack rolls against the creature have advantage.
Any attack that hits the creature is a critical hit if the attacker is within 5 feet of the creature.'),
('Petrified', 'A petrified creature is transformed, along with any nonmagical object it is wearing or carrying, into a solid inanimate substance (usually stone). Its weight increases by a factor of ten, and it ceases aging.
The creature is incapacitated (see the condition), can''t move or speak, and is unaware of its surroundings.
Attack rolls against the creature have advantage.
The creature automatically fails Strength and Dexterity saving throws.
The creature has resistance to all damage.
The creature is immune to poison and disease, although a poison or disease already in its system is suspended, not neutralized.'),
('Poisoned', 'A poisoned creature has disadvantage on attack rolls and ability checks.'),
('Prone', 'A prone creature''s only movement option is to crawl, unless it stands up and thereby ends the condition.
The creature has disadvantage on attack rolls.
An attack roll against the creature has advantage if the attacker is within 5 feet of the creature. Otherwise, the attack roll has disadvantage.'),
('Restrained', 'A restrained creature''s speed becomes 0, and it can''t benefit from any bonus to its speed.
Attack rolls against the creature have advantage, and the creature''s attack rolls have disadvantage.
The creature has disadvantage on Dexterity saving throws.'),
('Stunned', 'A stunned creature is incapacitated (see the condition), can''t move, and can speak only falteringly.
The creature automatically fails Strength and Dexterity saving throws.
Attack rolls against the creature have advantage.'),
('Unconscious', 'An unconscious creature is incapacitated (see the condition), can''t move or speak, and is unaware of its surroundings.
The creature drops whatever it''s holding and falls prone.
The creature automatically fails Strength and Dexterity saving throws.
Attack rolls against the creature have advantage.
Any attack that hits the creature is a critical hit if the attacker is within 5 feet of the creature.');

INSERT INTO Conditions (name, description, description_after) VALUES
('Exhaustion', 'Some special abilities and environmental hazards, such as starvation and the long-term effects of freezing or scorching temperatures, can lead to a special condition called exhaustion. Exhaustion is measured in six levels. An effect can give a creature one or more levels of exhaustion, as specified in the effect''s description.', 'If an already exhausted creature suffers another effect that causes exhaustion, its current level of exhaustion increases by the amount specified in the effect''s description.

A creature suffers the effect of its current level of exhaustion as well as all lower levels. For example, a creature suffering level 2 exhaustion has its speed halved and has disadvantage on ability checks.

An effect that removes exhaustion reduces its level as specified in the effect''s description, with all exhaustion effects ending if a creature''s exhaustion level is reduced below 1.

Finishing a long rest reduces a creature''s exhaustion level by 1, provided that the creature has also ingested some food and drink.');

CREATE TABLE ExhaustionLevels (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    level TEXT NOT NULL,
    effect TEXT NOT NULL
);

INSERT INTO ExhaustionLevels (level, effect) VALUES
('1', 'Disadvantage on ability checks'),
('2', 'Speed halved'),
('3', 'Disadvantage on attack rolls and saving throws'),
('4', 'Hit point maximum halved'),
('5', 'Speed reduced to 0'),
('6', 'Death');
