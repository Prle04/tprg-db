import os

filename = "test/commands.txt"
OS_DIR = "test/os"

if __name__ == "__main__":
    with open(filename, "w") as f:
        first_line = input("Enter first command to enter a db: ")
        f.write(first_line + "\n")
        for i in range(10000):
            key = f"key-{i:04d}"
            value = f"{i:04d}"
            f.write(f"set {key} {value}\n")
        
        for dir_entry in os.scandir(OS_DIR):
            if dir_entry.is_dir():
                folder_name = dir_entry.name
                folder_path = dir_entry.path

                for file_entry in os.scandir(folder_path):
                    if file_entry.is_file() and file_entry.name.endswith(".cpp"):
                        file_name = file_entry.name[:-4]  # remove '.cpp'
                        key = f"os-{folder_name}-{file_name}"
                        value = f'-f "test/os/{folder_name}/{file_entry.name}"'
                        f.write(f"set {key} {value}\n")
        f.write("logout\n")
        f.write("quit\n")
