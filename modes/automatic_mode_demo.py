def log_page1_values(self, data):
        # Page 1 values
        print("Page 1 - Automatic mode:")
        clone_miner_group = self.page1Layout.itemAt(0).widget()
        clone_miner_layout = clone_miner_group.layout()
        
        length_slider = clone_miner_layout.itemAt(0).layout().itemAt(1).widget()
        print(f"Minimal clone length in tokens: {length_slider.value()}")
        
        filtering_group = clone_miner_layout.itemAt(1).widget()
        filtering_layout = filtering_group.layout()
        
        convert_checkbox = filtering_layout.itemAt(0).widget()
        archetype_slider = filtering_layout.itemAt(1).layout().itemAt(1).widget()
        strict_filtering_checkbox = filtering_layout.itemAt(2).widget()
        
        print(f"Convert to DRL: {convert_checkbox.isChecked()}")
        print(f"Minimal archetype length in tokens: {archetype_slider.value()}")
        print(f"Strict small and overlapping duplicate filtering: {strict_filtering_checkbox.isChecked()}")
        print()