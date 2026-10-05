ALTER TABLE lessons ADD COLUMN title VARCHAR(255) NOT NULL DEFAULT 'Занятие';
UPDATE lessons SET title = COALESCE(NULLIF(notes, ''), 'Занятие');
