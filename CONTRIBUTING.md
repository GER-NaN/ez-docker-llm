Contributions are welcome that extend the core MVP to other platforms or hardware configurations on the condition the requirements and setup do not change. 

**Core Requirements**
- This repo remains a bare minimum setup for a local model running in docker that can be asked questions via an http request
- HTTP calls to the server require no additional information to ask for inference and the interface does not change from below

```
  //The interface that cannot change
  prompt(
    model,       
    prompt,
    context=null,
    format=null
  );
```

**Setup**
- Download and setup your model
- Configure Docker
- Start the server
- Ready!
