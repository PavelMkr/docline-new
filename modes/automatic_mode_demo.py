def page1_func(data):
    print("Page 1 - Automatic mode:")

    print(f"Minimal clone length in tokens: {data['length_slider']}")
    print(f"Convert to DRL: {data['convert_checkbox']}")
    print(f"Minimal archetype length in tokens: {data['archetype_slider']}")
    print(f"Strict small and overlapping duplicate filtering: {data['strict_filtering_checkbox']}")
    print()