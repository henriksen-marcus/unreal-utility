
<br>
<h2 align="center">unreal-utility</h2>
<br>
<p align="center">
 <img width="250px" src="https://github.com/henriksen-marcus/unreal-utility/assets/89453098/478031ad-2e43-459f-af83-f46a5efcfbfb"/>
</p>
<p align="center">Quickly delete temporary files, generate project files and compile the project in one click.</p>
<br>
<br>

## What is it?
unreal-utility is a program written in Go, deisgned to delete UE temporary files (Binaries, Intermediate etc.), automatically generate visual studio project files and finally compile the project for you.<br><br>
unreal-builder is a subset of unreal-utlity than only compiles the project. This allows you to compile without opening an IDE or UE first, saving RAM and drastically decreating compile time.

## :zap: How to use
Download the desired program from [releases](https://github.com/henriksen-marcus/unreal-utility/releases) and copy it to your unreal engine project directory. It should be on the same level as your `.uproject` file. **Make sure unreal engine and visual studio is closed before running, to avoid permission errors.** Then just run the exe to either clean + compile, or just compile depending on which program you downloaded.

## What do I need for it to work?
A working computer

## 🚑️ Troubleshooting
- If any problem occurs it will appear as red text in the terminal
- Most common issue is when you run the program too soon after closing unreal engine.
   - Try to wait at least 4 seconds after closing unreal engine, it takes some time for it to release all files from memory.

