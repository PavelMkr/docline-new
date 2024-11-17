import sys
from PyQt5.QtWidgets import (QFileDialog, QApplication, QWidget, QVBoxLayout, QHBoxLayout, QComboBox, 
                             QSlider, QLabel, QCheckBox, QLineEdit, QPushButton, QGroupBox)
from PyQt5.QtCore import Qt

class CloneMinerInterface(QWidget):
    def __init__(self):
        super().__init__()
        
        self.setWindowTitle("Clone Miner Interface")
        self.setFixedHeight(500)
        self.setFixedWidth(500)

        # Main layout
        main_layout = QVBoxLayout(self)

        # Analysis Method
        analysis_layout = QHBoxLayout()
        analysis_label = QLabel("Analysis method:")
        analysis_combo = QComboBox()
        analysis_combo.addItems(["Automatic mode", "Interactive mode", "Ngram Duplicate Finder", "Heuristic Ngram Finder"])
        analysis_layout.addWidget(analysis_label)
        analysis_layout.addWidget(analysis_combo)
        
        # Clone Miner settings group box
        clone_miner_group = QGroupBox("Clone Miner settings")
        clone_miner_layout = QVBoxLayout()

        # Minimal clone length slider
        length_layout = QHBoxLayout()
        length_label = QLabel("Minimal clone length in tokens")
        length_slider = QSlider(Qt.Horizontal)
        length_slider.setMinimum(1)
        length_slider.setMaximum(20)
        length_slider.setValue(5)
        length_slider.setTickPosition(QSlider.TicksBelow)
        length_slider.setTickInterval(1)
        length_value_label = QLabel("5")
        length_layout.addWidget(length_label)
        length_layout.addWidget(length_slider)
        length_layout.addWidget(length_value_label)

        # Update the value label when the slider changes
        length_slider.valueChanged.connect(lambda value: length_value_label.setText(str(value)))
        
        # Filtering options
        filtering_group = QGroupBox("Filtering")
        filtering_layout = QVBoxLayout()

        convert_checkbox = QCheckBox("Convert to DRL")
        
        archetype_layout = QHBoxLayout()
        archetype_label = QLabel("Minimal archetype length in tokens")
        archetype_slider = QSlider(Qt.Horizontal)
        archetype_slider.setMinimum(1)
        archetype_slider.setMaximum(20)
        archetype_slider.setValue(5)
        archetype_value_label = QLabel("5")
        archetype_layout.addWidget(archetype_label)
        archetype_layout.addWidget(archetype_slider)
        archetype_layout.addWidget(archetype_value_label)

        archetype_slider.valueChanged.connect(lambda value: archetype_value_label.setText(str(value)))
        
        strict_filtering_checkbox = QCheckBox("Strict small and overlapping duplicate filtering")
        
        # Add widgets to filtering layout
        filtering_layout.addWidget(convert_checkbox)
        filtering_layout.addLayout(archetype_layout)
        filtering_layout.addWidget(strict_filtering_checkbox)
        filtering_group.setLayout(filtering_layout)

        # Add everything to Clone Miner settings layout
        clone_miner_layout.addLayout(length_layout)
        clone_miner_layout.addWidget(filtering_group)
        clone_miner_group.setLayout(clone_miner_layout)

        # Source file
        source_file_layout = QHBoxLayout()
        source_file_label = QLabel("Source file")
        source_file_input = QLineEdit()
        source_file_button = QPushButton("...")
        source_file_layout.addWidget(source_file_label)
        source_file_layout.addWidget(source_file_input)
        source_file_layout.addWidget(source_file_button)
        
        # OK and Cancel buttons
        buttons_layout = QHBoxLayout()
        ok_button = QPushButton("OK")
        cancel_button = QPushButton("Cancel")
        buttons_layout.addStretch()
        buttons_layout.addWidget(ok_button)
        buttons_layout.addWidget(cancel_button)

        # Add layouts to main layout
        main_layout.addLayout(analysis_layout)
        main_layout.addWidget(clone_miner_group)
        main_layout.addLayout(source_file_layout)
        main_layout.addLayout(buttons_layout)

        self.setLayout(main_layout)

if __name__ == "__main__":
    app = QApplication(sys.argv)
    window = CloneMinerInterface()
    window.show()
    sys.exit(app.exec_())
