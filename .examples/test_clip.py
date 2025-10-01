#!/usr/bin/env python3
"""
Test CLIP installation and model loading
"""

import sys

def test_imports():
    """Test if required packages are installed"""
    print("Testing imports...")
    
    try:
        import torch
        print(f"✓ PyTorch {torch.__version__}")
    except ImportError:
        print("✗ PyTorch not installed")
        return False
    
    try:
        from PIL import Image
        print("✓ Pillow (PIL)")
    except ImportError:
        print("✗ Pillow not installed")
        return False
    
    try:
        import open_clip
        print(f"✓ OpenCLIP")
    except ImportError:
        print("✗ OpenCLIP not installed")
        return False
    
    return True

def test_model():
    """Test loading CLIP model"""
    print("\nTesting CLIP model loading...")
    
    try:
        import open_clip
        import torch
        
        model, _, preprocess = open_clip.create_model_and_transforms(
            'ViT-B-32',
            pretrained='openai',
            device='cpu'
        )
        
        print("✓ ViT-B/32 model loaded successfully")
        
        # Test text embedding
        tokenizer = open_clip.get_tokenizer('ViT-B-32')
        with torch.no_grad():
            text = tokenizer(["a photo of a cat"]).to('cpu')
            text_features = model.encode_text(text)
            
        print(f"✓ Text embedding shape: {text_features.shape}")
        print(f"✓ Embedding dimension: {text_features.shape[1]}")
        
        return True
        
    except Exception as e:
        print(f"✗ Error loading model: {e}")
        return False

def main():
    print("=== CLIP Installation Test ===\n")
    
    if not test_imports():
        print("\n❌ Some dependencies are missing")
        print("\nInstall with:")
        print("  pip install open_clip_torch pillow torch")
        sys.exit(1)
    
    if not test_model():
        print("\n❌ Model loading failed")
        sys.exit(1)
    
    print("\n✅ All tests passed! CLIP is ready to use.")
    print("\nYou can now use: same-same ingest -e clip images:./your_photos")

if __name__ == "__main__":
    main()
