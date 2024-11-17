import sys
from PyQt5.QtWidgets import (QApplication, QWidget, QVBoxLayout, QComboBox, QSlider, QCheckBox, 
                             QLabel, QPushButton, QHBoxLayout, QLineEdit, QFileDialog, QGroupBox)
from PyQt5.QtCore import Qt

class FuzzyHeatSettings(QWidget):
    def __init__(self):
        super().__init__()

        # Основные настройки окна
        self.setWindowTitle("Fuzzy Heat Settings")
        self.setFixedHeight(300)
        self.setFixedWidth(600)

        # Главный вертикальный макет
        main_layout = QVBoxLayout()

        # Выбор метода анализа
        analysis_method_layout = QHBoxLayout()
        analysis_method_label = QLabel("Analysis method:")
        self.analysis_method_combo = QComboBox()
        self.analysis_method_combo.addItems(["Automatic mode", "Interactive mode", "Ngram Duplicate Finder", "Heuristic Ngram Finder"])
        analysis_method_layout.addWidget(analysis_method_label)
        analysis_method_layout.addWidget(self.analysis_method_combo)
        main_layout.addLayout(analysis_method_layout)

        # Группа настроек Fuzzy Heat
        fuzzy_heat_group = QGroupBox("Fuzzy Heat Settings")
        fuzzy_heat_layout = QVBoxLayout()

        # Слайдер минимальной длины клона
        min_clone_layout = QHBoxLayout()
        min_clone_label = QLabel("Minimal clone length (number of tokens):")
        self.min_clone_slider = QSlider(Qt.Horizontal)
        self.min_clone_slider.setMinimum(1)
        self.min_clone_slider.setMaximum(20)
        self.min_clone_slider.setValue(5)
        self.min_clone_value = QLabel("5")
        self.min_clone_slider.valueChanged.connect(lambda value: self.min_clone_value.setText(str(value)))
        min_clone_layout.addWidget(min_clone_label)
        min_clone_layout.addWidget(self.min_clone_slider)
        min_clone_layout.addWidget(self.min_clone_value)

        # Слайдер максимальной длины клона
        max_clone_layout = QHBoxLayout()
        max_clone_label = QLabel("Maximal clone length (number of tokens):")
        self.max_clone_slider = QSlider(Qt.Horizontal)
        self.max_clone_slider.setMinimum(20)
        self.max_clone_slider.setMaximum(100)
        self.max_clone_slider.setValue(50)
        self.max_clone_value = QLabel("∞")
        self.max_clone_slider.valueChanged.connect(lambda value: self.max_clone_value.setText(str(value) if value < 100 else "∞"))
        max_clone_layout.addWidget(max_clone_label)
        max_clone_layout.addWidget(self.max_clone_slider)
        max_clone_layout.addWidget(self.max_clone_value)

        # Слайдер минимальной мощности группы
        min_group_layout = QHBoxLayout()
        min_group_label = QLabel("Minimal group power (number of clones):")
        self.min_group_slider = QSlider(Qt.Horizontal)
        self.min_group_slider.setMinimum(2)
        self.min_group_slider.setMaximum(10)
        self.min_group_slider.setValue(2)
        self.min_group_value = QLabel("2")
        self.min_group_slider.valueChanged.connect(lambda value: self.min_group_value.setText(str(value)))
        min_group_layout.addWidget(min_group_label)
        min_group_layout.addWidget(self.min_group_slider)
        min_group_layout.addWidget(self.min_group_value)

        # Чекбокс для расчета архетипов
        self.archetype_checkbox = QCheckBox("Extension point values")

        # Добавление всех элементов в макет Fuzzy Heat
        fuzzy_heat_layout.addLayout(min_clone_layout)
        fuzzy_heat_layout.addLayout(max_clone_layout)
        fuzzy_heat_layout.addLayout(min_group_layout)
        fuzzy_heat_layout.addWidget(self.archetype_checkbox)
        fuzzy_heat_group.setLayout(fuzzy_heat_layout)

        # Поле выбора файла
        file_layout = QHBoxLayout()
        file_label = QLabel("Source file:")
        self.file_path = QLineEdit()
        self.file_path.setPlaceholderText("Select a file...")
        self.file_button = QPushButton("...")
        self.file_button.clicked.connect(self.select_file)
        file_layout.addWidget(file_label)
        file_layout.addWidget(self.file_path)
        file_layout.addWidget(self.file_button)

        # Кнопки OK и Cancel
        button_layout = QHBoxLayout()
        ok_button = QPushButton("OK")
        cancel_button = QPushButton("Cancel")
        button_layout.addWidget(ok_button)
        button_layout.addWidget(cancel_button)

        # Добавление всех основных элементов в главный макет
        main_layout.addLayout(analysis_method_layout)
        main_layout.addWidget(fuzzy_heat_group)
        main_layout.addLayout(file_layout)
        main_layout.addLayout(button_layout)

        # Установка главного макета для окна
        self.setLayout(main_layout)

    def select_file(self):
        file_dialog = QFileDialog.getOpenFileName(self, "Select Source File", "", "All Files (*.*)")
        if file_dialog[0]:
            self.file_path.setText(file_dialog[0])

if __name__ == "__main__":
    app = QApplication(sys.argv)
    window = FuzzyHeatSettings()
    window.show()
    sys.exit(app.exec_())
