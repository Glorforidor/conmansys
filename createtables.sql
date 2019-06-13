-- Drop tables if they exists. Useful to flush data
DROP TABLE IF EXISTS conf_module_dependency;
DROP TABLE IF EXISTS conf_item_module;
DROP TABLE IF EXISTS conf_item;
DROP TABLE IF EXISTS conf_module;

-- Create conf_item table. 
-- Table can not consist of null data since it would not make since.
CREATE TABLE conf_item(
	conf_item_id SERIAL PRIMARY KEY,
	conf_item_value TEXT NOT NULL,
	conf_item_type TEXT NOT NULL,
	conf_item_version TEXT NOT NULL
);

-- Create conf_module table.
-- Table can not consist of null data since it would not make since.
CREATE TABLE conf_module(
	conf_module_id SERIAL PRIMARY KEY,
	conf_module_value TEXT NOT NULL,
	conf_module_version TEXT NOT NULL
);

-- Create conf_item_module table.
-- This table creates the connection between a item and module.
-- If a row from conf_item or conf_module is remove it will also remove the
-- connection.
CREATE TABLE conf_item_module(
	conf_item_module_id SERIAL PRIMARY KEY,
	conf_item_id INTEGER,
	conf_module_id INTEGER,
	FOREIGN KEY (conf_item_id) REFERENCES conf_item(conf_item_id) ON DELETE CASCADE,
	FOREIGN KEY (conf_module_id) REFERENCES conf_module(conf_module_id) ON DELETE CASCADE
);

-- Create conf_module_dependency table. 
-- This is a weak entity making it possible for modules depend on one another.
-- It checks for that dependent and dependee are not the same, since this will
-- make a module depend on itself.
CREATE TABLE conf_module_dependency(
	dependent int,
	dependee int,
    FOREIGN KEY (dependent) REFERENCES conf_module (conf_module_id),
    FOREIGN KEY (dependee) REFERENCES conf_module (conf_module_id),
	PRIMARY KEY (dependent, dependee),
    CONSTRAINT must_be_different CHECK (dependent != dependee)
);
