INSERT INTO conf_item (conf_item_value, conf_item_type, conf_item_version) VALUES
('tax_income_window', 'window', '1.0.0'),
('tax', 'domain', '1.0.0'),
('payment_window', 'window', '1.0.0'),
('payment', 'domain', '1.0.0'),
('refund_window', 'window', '2.0.0'),
('refund', 'domain', '0.0.1'),
('management_tax_window', 'window', '2.0.0'),
('mangement', 'domain', '0.0.2');

INSERT INTO conf_module (conf_module_value, conf_module_version) VALUES
('A', '0.0.10'),
('B', '0.0.11'),
('C', '0.0.12'),
('D', '0.0.13'),
('E', '0.0.14'),
('F', '0.0.15');

INSERT INTO conf_item_module (conf_item_id, conf_module_id) VALUES
(1, 1),
(2, 1),
(3, 2),
(4, 2),
(5, 3),
(6, 3),
(7, 4),
(8, 4);

INSERT INTO conf_module_dependency (dependent, dependee) VALUES
(4, 1),
(4, 2),
(4, 3),
(5, 4),
(6, 1),
(6, 5);
