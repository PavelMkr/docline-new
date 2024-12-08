import sys, requests, json
import subprocess, time, signal, sys
from PyQt5.QtWidgets import (QApplication, QWidget, QVBoxLayout, QComboBox, QPushButton, QLabel, 
                             QFileDialog, QStackedLayout, QFormLayout, QSlider, QGroupBox, QHBoxLayout, 
                             QCheckBox)
from PyQt5.QtCore import Qt


# from modes.automatic_mode_demo import page1_func
# from modes.interactive_mode_demo import page2_func
# from modes.ngram_duplicate_finder_demo import page3_func
# from modes.heuristic_ngram_finder_demo import page4_func

class DocLine(QWidget):
    def __init__(self):
        super().__init__()
        #server starts
        self.go_server_process = subprocess.Popen(["go","run","modes/server.go"]) # file version
        #self.go_server_process = subprocess.Popen(["./modes/server"]) # compile version
        signal.signal(signal.SIGINT, self.stop_server)
        signal.signal(signal.SIGTERM, self.stop_server)
        # window settings
        self.setWindowTitle("DocLine demo")
        self.setFixedHeight(500)
        self.setFixedWidth(500)
        # Create a top-level layout
        layout = QVBoxLayout()
        self.setLayout(layout)
        
        # Create and connect the combo box to switch between pages
        self.pageCombo = QComboBox()
        self.pageCombo.addItems(["Automatic mode", "Interactive mode", "Ngram Duplicate Finder", "Heuristic Ngram Finder"])
        self.pageCombo.activated.connect(self.switchPage)

        # Choose file button
        self.file_button = QPushButton("Choose File", self)
        self.file_button.clicked.connect(self.open_file_dialog)

        # File path label
        self.file_path_label = QLabel("No file", self)

        # Start Process button
        self.start_button = QPushButton('Start', self)
        self.start_button.clicked.connect(self.start_function)

        # Create the stacked layout
        self.stackedLayout = QStackedLayout()

        # Page 1 - Automatic mode
        self.page1 = QWidget()
        self.page1Layout = QVBoxLayout()

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

        # Add Clone Miner group to page1 layout
        self.page1Layout.addWidget(clone_miner_group)
        self.page1.setLayout(self.page1Layout)
        
        # Add page1 to stacked layout
        self.stackedLayout.addWidget(self.page1)

        # Page 2 - Interactive mode (Fuzzy Heat Settings)
        self.page2 = QWidget()
        self.page2Layout = QVBoxLayout()

        # Fuzzy Heat Settings group box
        fuzzy_heat_group = QGroupBox("Fuzzy Heat Settings")
        fuzzy_heat_layout = QVBoxLayout()

        # Minimal clone length slider
        min_clone_layout = QHBoxLayout()
        min_clone_label = QLabel("Minimal clone length (number of tokens):")
        min_clone_slider = QSlider(Qt.Horizontal)
        min_clone_slider.setMinimum(1)
        min_clone_slider.setMaximum(20)
        min_clone_slider.setValue(5)
        min_clone_value = QLabel("5")
        min_clone_slider.valueChanged.connect(lambda value: min_clone_value.setText(str(value)))
        min_clone_layout.addWidget(min_clone_label)
        min_clone_layout.addWidget(min_clone_slider)
        min_clone_layout.addWidget(min_clone_value)

        # Maximal clone length slider
        max_clone_layout = QHBoxLayout()
        max_clone_label = QLabel("Maximal clone length (number of tokens):")
        max_clone_slider = QSlider(Qt.Horizontal)
        max_clone_slider.setMinimum(2)
        max_clone_slider.setMaximum(200)
        max_clone_slider.setValue(50)
        max_clone_value = QLabel("50")
        max_clone_slider.valueChanged.connect(lambda value: max_clone_value.setText(str(value) if value < 100 else "∞"))
        max_clone_layout.addWidget(max_clone_label)
        max_clone_layout.addWidget(max_clone_slider)
        max_clone_layout.addWidget(max_clone_value)

        # Minimal group power slider
        min_group_layout = QHBoxLayout()
        min_group_label = QLabel("Minimal group power (number of clones):")
        min_group_slider = QSlider(Qt.Horizontal)
        min_group_slider.setMinimum(2)
        min_group_slider.setMaximum(10)
        min_group_slider.setValue(2)
        min_group_value = QLabel("2")
        min_group_slider.valueChanged.connect(lambda value: min_group_value.setText(str(value)))
        min_group_layout.addWidget(min_group_label)
        min_group_layout.addWidget(min_group_slider)
        min_group_layout.addWidget(min_group_value)

        # Extension point values checkbox
        extension_checkbox = QCheckBox("Extension point values")

        # Add all elements to fuzzy heat layout
        fuzzy_heat_layout.addLayout(min_clone_layout)
        fuzzy_heat_layout.addLayout(max_clone_layout)
        fuzzy_heat_layout.addLayout(min_group_layout)
        fuzzy_heat_layout.addWidget(extension_checkbox)
        fuzzy_heat_group.setLayout(fuzzy_heat_layout)

        # Add fuzzy heat group to page2 layout
        self.page2Layout.addWidget(fuzzy_heat_group)
        self.page2.setLayout(self.page2Layout)
        
        # Add page2 to stacked layout
        self.stackedLayout.addWidget(self.page2)

        # Page 3 - Ngram Duplicate Finder
        self.page3 = QWidget()
        self.page3Layout = QFormLayout()

        # Fuzzy Finder Settings group box
        fuzzy_finder_group = QGroupBox("Fuzzy Finder Settings")
        fuzzy_finder_layout = QVBoxLayout()

        # Minimal clone length slider
        min_clone_layout = QHBoxLayout()
        min_clone_label = QLabel("Minimal clone length (number of tokens):")
        min_clone_slider = QSlider(Qt.Horizontal)
        min_clone_slider.setMinimum(1)
        min_clone_slider.setMaximum(20)
        min_clone_slider.setValue(5)
        min_clone_value = QLabel("5")
        min_clone_slider.valueChanged.connect(lambda value: min_clone_value.setText(str(value)))
        min_clone_layout.addWidget(min_clone_label)
        min_clone_layout.addWidget(min_clone_slider)
        min_clone_layout.addWidget(min_clone_value)

        # Maximal edit distance slider
        max_edit_layout = QHBoxLayout()
        max_edit_label = QLabel("Maximal edit distance (Levenshtein):")
        max_edit_slider = QSlider(Qt.Horizontal)
        max_edit_slider.setMinimum(2)
        max_edit_slider.setMaximum(200)
        max_edit_slider.setValue(50)
        max_edit_value = QLabel("50")
        max_edit_slider.valueChanged.connect(lambda value: max_edit_value.setText(str(value) if value < 100 else "∞"))
        max_edit_layout.addWidget(max_edit_label)
        max_edit_layout.addWidget(max_edit_slider)
        max_edit_layout.addWidget(max_edit_value)

        # Minimal group power (number of clones) slider
        max_fuzzy_layout = QHBoxLayout()
        max_fuzzy_label = QLabel("Minimal group power (number of clones):")
        max_fuzzy_slider = QSlider(Qt.Horizontal)
        max_fuzzy_slider.setMinimum(2)
        max_fuzzy_slider.setMaximum(10)
        max_fuzzy_slider.setValue(2)
        max_fuzzy_value = QLabel("2")
        max_fuzzy_slider.valueChanged.connect(lambda value: max_fuzzy_value.setText(str(value)))
        max_fuzzy_layout.addWidget(max_fuzzy_label)
        max_fuzzy_layout.addWidget(max_fuzzy_slider)
        max_fuzzy_layout.addWidget(max_fuzzy_value)

        # Source language selection
        language_layout = QHBoxLayout()
        language_label = QLabel("Source Language:")
        self.source_language = QComboBox()
        self.source_language.addItems(["English", "Russian"])
        language_layout.addWidget(language_label)
        language_layout.addWidget(self.source_language)

        # Add elements to fuzzy finder layout
        fuzzy_finder_layout.addLayout(min_clone_layout)
        fuzzy_finder_layout.addLayout(max_edit_layout)
        fuzzy_finder_layout.addLayout(max_fuzzy_layout)
        fuzzy_finder_layout.addLayout(language_layout)
        fuzzy_finder_group.setLayout(fuzzy_finder_layout)

        # Add fuzzy finder group to page3 layout
        self.page3Layout.addWidget(fuzzy_finder_group)
        self.page3.setLayout(self.page3Layout)
        self.stackedLayout.addWidget(self.page3)

        # Page 4 - Heuristic Ngram Finder
        self.page4 = QWidget()
        self.page4Layout = QFormLayout()

        # Heuristic Duplicate Finder group box
        heurestic_duplicate_group = QGroupBox("Heuristic Duplicate Finder")
        heurestic_duplicate_layout = QVBoxLayout()

        # Extension point values checkbox
        extention_point_checkbox = QCheckBox("Extension point values")
        heurestic_duplicate_layout.addWidget(extention_point_checkbox)

        # Set layout for the group box
        heurestic_duplicate_group.setLayout(heurestic_duplicate_layout)

        # Add the group box to the page layout
        self.page4Layout.addWidget(heurestic_duplicate_group)

        # Set the layout for the page and add it to the stacked layout
        self.page4.setLayout(self.page4Layout)
        self.stackedLayout.addWidget(self.page4)

        # Add the combo box and the stacked layout to the top-level layout
        layout.addWidget(self.pageCombo)
        layout.addLayout(self.stackedLayout)
        layout.addWidget(self.file_button)
        layout.addWidget(self.file_path_label)
        layout.addWidget(self.start_button)

    def switchPage(self):
        self.stackedLayout.setCurrentIndex(self.pageCombo.currentIndex())
            
    def open_file_dialog(self):
        options = QFileDialog.Options()
        file_name, _ = QFileDialog.getOpenFileName(self, "Выберите файл", "", "Все файлы (*);;Текстовые файлы (*.txt)", options=options)
        if file_name:
            self.file_path_label.setText(file_name)
            print(f"Chosen file: {file_name}")

    def start_function(self):
        current_index = self.stackedLayout.currentIndex()
        
        if current_index == 0:  # Page 1 - Automatic mode
            clone_miner_group = self.page1Layout.itemAt(0).widget()
            clone_miner_layout = clone_miner_group.layout()
            
            length_slider = clone_miner_layout.itemAt(0).layout().itemAt(1).widget()

            filtering_group = clone_miner_layout.itemAt(1).widget()
            filtering_layout = filtering_group.layout()

            convert_checkbox = filtering_layout.itemAt(0).widget()
            archetype_slider = filtering_layout.itemAt(1).layout().itemAt(1).widget()
            strict_filtering_checkbox = filtering_layout.itemAt(2).widget()

            data = {
                'length_slider': length_slider.value(),
                'convert_checkbox':convert_checkbox.isChecked(),
                'archetype_slider': archetype_slider.value(),
                'strict_filtering_checkbox': strict_filtering_checkbox.isChecked()
            }
            response = requests.post("http://localhost:8080/automatic_mode", json=data)
            print(f"Automatic Mode Response Status: {response.status_code}")
            #page1_func(data)
        elif current_index == 1:  # Page 2 - Interactive mode
            fuzzy_heat_group = self.page2Layout.itemAt(0).widget()
            fuzzy_heat_layout = fuzzy_heat_group.layout()
            
            min_clone_slider = fuzzy_heat_layout.itemAt(0).layout().itemAt(1).widget()
            max_clone_slider = fuzzy_heat_layout.itemAt(1).layout().itemAt(1).widget()
            min_group_slider = fuzzy_heat_layout.itemAt(2).layout().itemAt(1).widget()
            extension_checkbox = fuzzy_heat_layout.itemAt(3).widget()

            data = {
                'min_clone_slider': min_clone_slider.value(),
                'max_clone_slider':max_clone_slider.value(),
                'min_group_slider': min_group_slider.value(),
                'extension_checkbox': extension_checkbox.isChecked()
            }
            response = requests.post("http://localhost:8080/interactive_mode", json=data)
            print(f"Interactive Mode Response Status: {response.status_code}")
            #page2_func(data)
        elif current_index == 2:  # Page 3 - Ngram Duplicate Finder
            fuzzy_finder_group = self.page3Layout.itemAt(0).widget()
            fuzzy_finder_layout = fuzzy_finder_group.layout()
            
            min_clone_slider = fuzzy_finder_layout.itemAt(0).layout().itemAt(1).widget()
            max_edit_slider = fuzzy_finder_layout.itemAt(1).layout().itemAt(1).widget()
            max_fuzzy_slider = fuzzy_finder_layout.itemAt(2).layout().itemAt(1).widget()
            source_language = fuzzy_finder_layout.itemAt(3).layout().itemAt(1).widget()

            data = {
                'min_clone_slider': min_clone_slider.value(),
                'max_edit_slider': max_edit_slider.value(),
                'max_fuzzy_slider': max_fuzzy_slider.value(),
                'source_language': source_language.currentText()
            }
            response = requests.post("http://localhost:8080/ngram_finder", json=data)
            print(f"Ngram Duplicate Response Status: {response.status_code}" )
            #page3_func(data)
        elif current_index == 3:  # Page 4 - Heuristic Ngram Finder
            heurestic_duplicate_group = self.page4Layout.itemAt(0).widget()
            heurestic_duplicate_layout = heurestic_duplicate_group.layout()
        
            extention_point_checkbox = heurestic_duplicate_layout.itemAt(0).widget()

            data = {
                'extention_point_checkbox': extention_point_checkbox.isChecked()
            }
            response = requests.post("http://localhost:8080/heuristic_finder", json=data)
            print(f"Heuristic Ngram Response Status: {response.status_code}" )
            #page4_func(data)  

    def closeEvent(self, close):
        self.stop_server()
        close.accept()

    def stop_server(self):
        print("Stop Go server")
        self.go_server_process.terminate()
        self.go_server_process.wait()
        #sys.exit(0) 


if __name__ == "__main__":
    app = QApplication(sys.argv)
    window = DocLine()
    window.show()
    sys.exit(app.exec_())
